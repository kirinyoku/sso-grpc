// Package config provides functionality for loading application configuration
package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the application configuration structure.
// It holds general settings and nested GRPC configuration.
type Config struct {
	Env         string        `yaml:"env" env-default:"local"`          // Application environment (e.g., local, dev, prod)
	StoragePath string        `yaml:"storage_path" env-required:"true"` // Path to the storage or database file
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`    // Time-to-live for access tokens
	GRPC        GRPC          `yaml:"grpc"`                             // GRPC server-related settings
}

// GRPC holds configuration values related to the GRPC server.
type GRPC struct {
	Port    int           `yaml:"port" env-required:"true"` // Port on which the GRPC server runs
	Timeout time.Duration `yaml:"timeout" env-default:"1h"` // Request timeout for GRPC server
}

// MustLoad loads the application configuration from a YAML file
// whose path is provided via the --config flag or CONFIG_PATH env var.
// It panics if the configuration cannot be loaded or the file is invalid.
func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is not specified")
	}

	return MustLoadByPath(configPath)
}

func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to load config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath retrieves the path to the configuration file.
// It first checks the --config command-line flag,
// then falls back to the CONFIG_PATH environment variable.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "Path to the config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
