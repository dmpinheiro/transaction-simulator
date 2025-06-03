package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func New() *viper.Viper {
	configFile := "config.yaml"

	viper := viper.New()
	viper.SetDefault("database.file", "simulator.db")
	viper.SetDefault("app.port", 3000)
	viper.SetConfigName("config") // name of config file (without extension)
        viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error loading configuration : %+v", err))
	}

	return viper
}