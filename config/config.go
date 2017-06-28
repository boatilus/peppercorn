package config

import (
	"log"
	"os/user"

	"github.com/spf13/viper"
)

const configName = "peppercorn"
const configPath = "."

func init() {
	viper.SetConfigName(configName)

	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(u.HomeDir)
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		// We needn't panic or anything here, since we'll create the config with functional defaults.
		log.Print("config: Could not read config file")
	}

	// Merely return and skip configuring the Sentry hook if no Sentry DSN specified in the config.
	dsn := viper.GetString("sentry.dsn")
	if dsn == "" {
		return
	}
}

func GetString(key string) string {
	return viper.GetString(key)
}
