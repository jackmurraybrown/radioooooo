package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL      string
	Port             string
	JWTSecret        string
	AllowedOrigins   []string
	LiquidsoapSocket string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             "8080",
		AllowedOrigins:   []string{"http://localhost:5173"},
		LiquidsoapSocket: "/var/run/liquidsoap/radio.sock",
	}
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if p := os.Getenv("PORT"); p != "" {
		cfg.Port = p
	}
	if o := os.Getenv("ALLOWED_ORIGINS"); o != "" {
		cfg.AllowedOrigins = strings.Split(o, ",")
	}
	if s := os.Getenv("LIQUIDSOAP_SOCKET"); s != "" {
		cfg.LiquidsoapSocket = s
	}
	return cfg, nil
}
