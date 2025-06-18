// pkg/config/config.go
package config

import (
	"os"
	"strings"
)

type Config struct {
	Port           string
	Environment    string
	DatabaseURL    string
	RedisAddr      string
	RedisPassword  string
	AllowedOrigins []string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/mindkeep?sslmode=disable"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
