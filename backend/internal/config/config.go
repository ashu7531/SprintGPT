// Package config handles application configuration loaded from environment variables.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	ServerPort string
	CORSOrigin string
	RedisURL   string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	// Try to load .env file; it's okay if it doesn't exist (e.g., in production)
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		CORSOrigin: getEnv("CORS_ORIGIN", "http://localhost:5173"),
		RedisURL:   getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
