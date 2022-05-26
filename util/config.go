package util

import "github.com/spf13/viper"

type Config struct {
	DB_Driver      string `mapstructure:"DB_DRIVER"`
	DB_Source      string `mapstructure:"DB_SOURCE"`
	Server_Address string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	var config Config
	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	return config, err
}