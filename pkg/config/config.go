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

	viper.SetDefault("appRepo.reconciliationInterval", 1*time.Hour)
	viper.SetDefault("configRepo.branchName", "main")
	viper.SetDefault("configRepo.reconciliationInterval", 30*time.Minute)
	viper.SetDefault("ssh.keyDir", "/var/lib/recoon")
	viper.SetDefault("store.databaseFile", "/var/lib/recoon/bbolt.db")
	viper.SetDefault("store.gitDir", "/var/lib/recoon/repos")
	viper.SetDefault("ui.port", 3680)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	return viper.GetViper(), nil
}
