package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`

	Mongo struct {
		URI     string `mapstructure:"uri"`
		DB      string `mapstructure:"db"`
		Timeout int    `mapstructure:"timeout"`
	} `mapstructure:"mongo"`

	App struct {
		PostsPerPage int `mapstructure:"posts_per_page"`
	} `mapstructure:"app"`
}

func Load() (Config, error) {
	var config Config

	v := viper.New()

	defaultHost := "localhost"
	if isRunningInContainer() {
		defaultHost = "0.0.0.0"
	}

	v.SetDefault("server.host", defaultHost)
	v.SetDefault("server.port", "8080")
	v.SetDefault("mongo.uri", "mongodb://localhost:27017")
	v.SetDefault("mongo.db", "news_app")
	v.SetDefault("mongo.timeout", 10)
	v.SetDefault("app.posts_per_page", 12)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := v.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return config, nil
}

// isRunningInContainer detects if the app is running inside a container
func isRunningInContainer() bool {
	if os.Getenv("CONTAINER") != "" {
		return true
	}

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	if os.Getenv("SERVER_HOST") != "" {
		return true
	}

	return false
}
