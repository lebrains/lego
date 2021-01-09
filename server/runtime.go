package server

import (
	"github.com/spf13/cobra"

	"github.com/vseinstrumentiru/lego/v2/common/cast"
	"github.com/vseinstrumentiru/lego/v2/common/set"
	"github.com/vseinstrumentiru/lego/v2/internal/env"
)

const (
	defaultEnvPath = "app"
)

func newRuntime(opts []Option) *Runtime {
	runtime := &Runtime{
		opts: cast.NewCastableRWSet(set.NewSimple()),
		cmd: &cobra.Command{
			Use:    "lego",
			Hidden: true,
		},
	}

	for _, opt := range opts {
		opt(runtime)
	}

	return runtime
}

type Runtime struct {
	opts cast.CastableRWSet
	cfg  interface{}
	cmd  *cobra.Command
}

func (r *Runtime) On(key string, callback interface{}) (ok bool) {
	return cast.OnCheck(r.opts, key, callback)
}

func (r *Runtime) onConfig(cb func(config interface{})) {
	if r.cfg != nil {
		cb(r.cfg)
	}
}

func (r *Runtime) Is(key string) bool {
	ok, err := r.opts.GetBool(key)

	return ok && err == nil
}

func (r *Runtime) Not(key string) bool {
	ok, err := r.opts.GetBool(key)

	return !ok || err != nil
}

func (r *Runtime) newEnv() env.Env {
	path := defaultEnvPath

	r.On(optEnvPath, func(newPath string) {
		path = newPath
	})

	if r.cfg == nil {
		return env.NewNoConfigEnv(r.cmd.Flags(), path)
	}

	return env.NewConfigEnv(r.cmd.Flags(), path)
}
