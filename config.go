package main

import (
	"flag"
	"github.com/spf13/viper"
	"time"
)

// setDefaultConfig is called once per run at the very beginning.
// This ensures that critical configuration values are set.
func setDefaultConfig() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 80)
	viper.SetDefault("server.max_connections", 16)
	viper.SetDefault("server.timeout", time.Minute)
	viper.SetDefault("upstream.host", "127.0.0.1")
	viper.SetDefault("upstream.port", 8080)
}

// loadConfig must be called in main to initialize application's configuration
func loadConfig() error {
	setDefaultConfig()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	configPath := flag.String("config", "", "Path to config.yaml")
	flag.Parse()
	if *configPath != "" {
		viper.SetConfigFile(*configPath)
	} else {
		viper.AddConfigPath("/etc/authentication_proxy/")
		viper.AddConfigPath(".")
	}

	return viper.ReadInConfig()
}
