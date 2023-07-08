package config

import (
	"github.com/spf13/viper"
	"time"
)

type Getter interface {
	GetString(key string) string
	GetInt(key string) int
	GetDuration(key string) time.Duration
	Sub(key string) *viper.Viper
}

func Setup() (Getter, error) {
	viper.SetConfigName("recooncfg")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/recoon")
	viper.AddConfigPath("$HOME/.recoon")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	return viper.GetViper(), nil
}
