// Package config loads taskflow-go's runtime configuration from environment
// variables. There is no DI framework: main wires concrete values through
// plain constructor functions, and config is just the first one of those.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every environment-provided setting the service needs.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

// Load reads configuration from the environment, applying sane defaults for
// local development and failing fast if anything required is missing.
func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}

	ttlMinutes := getEnv("JWT_TTL_MINUTES", "60")
	minutes, err := strconv.Atoi(ttlMinutes)
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_TTL_MINUTES %q: %w", ttlMinutes, err)
	}
	cfg.JWTTTL = time.Duration(minutes) * time.Minute

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
