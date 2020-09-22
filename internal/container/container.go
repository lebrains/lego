package container

import (
	"fmt"
	"reflect"
	"strings"

	"emperror.dev/errors"
	"go.uber.org/dig"

	"github.com/vseinstrumentiru/lego/inject"
)

func New() *container {
	instance := dig.New()

	c := &container{
		di: instance,
	}

	return c
}

type container struct {
	di *dig.Container
}

func (c *container) Register(constructor inject.Constructor, options ...inject.RegisterOption) error {
	return errors.WithStack(c.di.Provide(constructor, options...))
}

func (c *container) Instance(instance interface{}, options ...inject.RegisterOption) error {
	t := reflect.ValueOf(instance)

	if err := checkStruct(t); err != nil {
		return errors.WithStack(err)
	}

	funcType := reflect.FuncOf(nil, []reflect.Type{t.Type()}, false)
	f := reflect.MakeFunc(funcType, instanceFn(t))

	return c.Register(f.Interface(), options...)
}

func (c *container) Execute(function inject.Invocation) error {
	return errors.WithStack(c.di.Invoke(function))
}

func (c *container) Make(i inject.Interface) error {
	val := reflect.ValueOf(i)

	if err := checkStruct(val); err != nil {
		return err
	}

	if appWithProviders, ok := i.(interface{ Providers() []interface{} }); ok {
		constructors := appWithProviders.Providers()
		for i := 0; i < len(constructors); i++ {
			if err := c.Register(constructors[i]); err != nil {
				return err
			}
		}
	}

	t := val.Type()
	var configurations []interface{}
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		if m.Name == "Provide" || strings.HasPrefix(m.Name, "Provide") {
			constructor := val.MethodByName(m.Name).Interface()
			if err := c.Register(constructor); err != nil {
				return err
			}
		} else if m.Name == "Configure" || strings.HasPrefix(m.Name, "Configure") {
			configurations = append(configurations, val.MethodByName(m.Name).Interface())
		}
	}

	if err := c.resolve(i); err != nil {
		return errors.WithStack(err)
	}

	for i := 0; i < len(configurations); i++ {
		if err := c.Execute(configurations[i]); err != nil {
			return err
		}
	}

	return nil
}

func (c *container) resolve(i inject.Interface) error {
	val := reflect.ValueOf(i)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var inFields, outFields []reflect.StructField
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if alias, ok := field.Tag.Lookup("name"); !ok {
			inFields = append(inFields, reflect.StructField{
				Name: field.Name,
				Type: field.Type,
				Tag:  reflect.StructTag(fmt.Sprintf(`name:"%s"`, alias)),
			})
		} else {
			inFields = append(inFields, reflect.StructField{
				Name: field.Name,
				Type: field.Type,
				Tag:  `optional:"true"`,
			})
		}

		outFields = append(outFields, reflect.StructField{
			Name: field.Name,
			Type: field.Type,
		})
	}

	if len(inFields) == 0 {
		return nil
	}

	inFields = append(inFields, reflect.StructField{
		Name:      "In",
		Type:      reflect.TypeOf(dig.In{}),
		Anonymous: true,
	})

	outFields = append(outFields, reflect.StructField{
		Name:      "Out",
		Type:      reflect.TypeOf(dig.Out{}),
		Anonymous: true,
	})

	in := reflect.Indirect(reflect.New(reflect.StructOf(inFields)))
	out := reflect.Indirect(reflect.New(reflect.StructOf(outFields)))
	fn := reflect.MakeFunc(reflect.FuncOf([]reflect.Type{in.Type()}, []reflect.Type{out.Type()}, false), func(args []reflect.Value) (results []reflect.Value) {
		arg := args[0]

		for i := 0; i < arg.Type().NumField(); i++ {
			field := arg.Type().Field(i)
			if field.Anonymous {
				continue
			}
			out.FieldByName(field.Name).Set(arg.FieldByName(field.Name))
		}

		return []reflect.Value{out}
	})

	if err := c.di.Invoke(fn.Interface()); err != nil {
		return err
	}

	for i := 0; i < out.Type().NumField(); i++ {
		field := out.Type().Field(i)
		if field.Anonymous {
			continue
		}
		val.FieldByName(field.Name).Set(out.FieldByName(field.Name))
	}

	return nil
}

func instanceFn(i reflect.Value) func(args []reflect.Value) []reflect.Value {
	return func(args []reflect.Value) []reflect.Value {
		return []reflect.Value{i}
	}
}

func resolveFn(i interface{}) interface{} {
	var outFields []reflect.StructField
	val := reflect.ValueOf(i)
	t := val.Type()

	outFields = append(outFields, reflect.StructField{
		Type:      reflect.TypeOf(dig.Out{}),
		Anonymous: true,
	})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		outFields = append(outFields, reflect.StructField{
			Name: field.Name,
			Type: field.Type,
		})
	}

	result := reflect.New(reflect.StructOf(outFields))

	for i := 0; i < result.NumField(); i++ {
		field := result.Field(i)
		field.Set(val.Field(i))
	}

	return result
}

func checkStruct(t reflect.Value) error {
	if t.Kind() == reflect.Ptr {
		if t.IsNil() || !t.IsValid() {
			return errors.New("nil instance presented")
		}

		ti := reflect.Indirect(t)

		if ti.Kind() != reflect.Struct {
			return errors.New("instance must be struct or non-nil pointer to struct")
		}
	} else if t.Kind() != reflect.Struct {
		return errors.New("instance must be struct or non-nil pointer to struct")
	}

	return nil
}
