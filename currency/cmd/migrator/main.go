package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/vctrl/currency-service/currency/internal/config"
	"github.com/vctrl/currency-service/currency/internal/migrations"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

func run() error {
	configPath := flag.String("config", "./config", "path to the config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := migrations.RunPgMigrations(cfg.Database); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

type appConfig struct {
	Database config.DatabaseConfig `mapstructure:"database"`
}

func loadConfig(path string) (appConfig, error) {
	var config appConfig

	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	return config, nil
}
