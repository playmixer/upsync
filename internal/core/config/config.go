package config

import (
	"errors"
	"fmt"
	"os"
	"upsync/internal/core/upsync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config - конфиг сервиса.
type Config struct {
	LogLevel string `env:"LOG_LEVEL"`
	LogPath  string `env:"LOG_PATH"`
	Upsync   upsync.Config
}

// Init - инициализация конфига.
func Init() (*Config, error) {
	cfg := &Config{
		Upsync: upsync.Config{},
	}

	cfgFile := ".env"

	err := godotenv.Load(cfgFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed load env: %w", err)
	}

	err = env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed parse environments: %w", err)
	}

	return cfg, nil
}
