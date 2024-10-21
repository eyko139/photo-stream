package env

import (
	"github.com/spf13/viper"
)

type Env struct {
	PrismURL string `mapstructure:"PRISM_URL"`
	PrismAuthToken    string `mapstructure:"PRISM_AUTH_TOKEN"`
}

func NewEnv() *Env {
	env := Env{}

	viper.BindEnv("PRISM_URL")
	viper.SetDefault("PRISM_URL", "")
	viper.BindEnv("PRISM_AUTH_TOKEN")
	viper.SetDefault("PRISM_AUTH_TOKEN", "")

	env.PrismURL = viper.GetString("PRISM_URL")
	env.PrismAuthToken = viper.GetString("PRISM_AUTH_TOKEN")

	return &env
}
