package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DB_Driver      string `mapstructure:"DB_DRIVER"`
	DB_Source      string `mapstructure:"DB_SOURCE"`
	Server_Address string `mapstructure:"SERVER_ADDRESS"`
	Access_Token string `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	Token_Duration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}