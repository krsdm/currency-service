package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	ServerPort string `mapstructure:"server_port"`
}

type JWTConfig struct {
	JWTSecret          string `mapstructure:"jwt_secret"`
	JWTLifetimeSeconds int    `mapstructure:"jwt_lifetime_seconds"`
}

type Config struct {
	Service ServerConfig `mapstructure:"service"`
	JWT     JWTConfig    `mapstructure:"jwt"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	return cfg, nil
}
