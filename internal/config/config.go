package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	cfg := &Config{Port: "8080"}
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if p := os.Getenv("PORT"); p != "" {
		cfg.Port = p
	}
	return cfg, nil
}
