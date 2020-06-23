package http

import (
	"emperror.dev/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	lego2 "github.com/vseinstrumentiru/lego/internal/lego"
)

type Config struct {
	lego2.WithSwitch `mapstructure:",squash"`
	IsPublic         bool

	Port int
}

func (c Config) SetDefaults(env *viper.Viper, flag *pflag.FlagSet) {
	flag.Int("http-port", 8080, "HTTP server port")
	_ = env.BindPFlag("srv.http.port", flag.Lookup("http-port"))
	env.SetDefault("srv.http.port", 8080)
}

func (c Config) Validate() (err error) {
	if c.Enabled {
		if c.Port == 0 {
			err = errors.Append(err, errors.New("srv.http.port is required"))
		}
	}

	return
}