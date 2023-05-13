package config

import "github.com/spf13/viper"

type Config struct {
	User     string `mapstructure:"USER"`
	Pass     string `mapstructure:"PASSWORD"`
	BaseUrl  string `mapstructure:"BASE_URL"`
	Services string `mapstructure:"SERVICES"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("properties")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	viper.Unmarshal(&config)
	return
}
