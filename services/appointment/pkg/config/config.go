package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPPort     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:     envOrDefault("HTTP_PORT", "8082"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if value := os.Getenv("READ_TIMEOUT"); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse READ_TIMEOUT: %w", err)
		}
		cfg.ReadTimeout = duration
	}

	if value := os.Getenv("WRITE_TIMEOUT"); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse WRITE_TIMEOUT: %w", err)
		}
		cfg.WriteTimeout = duration
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
