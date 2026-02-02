package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Admin    AdminConfig
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port      int
	Host      string
	LocalOnly bool
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Path string
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// AdminConfig holds default admin configuration.
type AdminConfig struct {
	Username string
	Password string
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:      getEnvInt("WG_MGT_PORT", 8080),
			Host:      getEnv("WG_MGT_HOST", "127.0.0.1"),
			LocalOnly: getEnvBool("WG_MGT_LOCAL_ONLY", true),
		},
		Database: DatabaseConfig{
			Path: getEnv("WG_MGT_DB_PATH", "wg-mgt.db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("WG_MGT_JWT_SECRET", generateDefaultSecret()),
			Expiration: getEnvDuration("WG_MGT_JWT_EXPIRATION", 24*time.Hour),
		},
		Admin: AdminConfig{
			Username: getEnv("WG_MGT_ADMIN_USER", ""),
			Password: getEnv("WG_MGT_ADMIN_PASS", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func generateDefaultSecret() string {
	return "change-me-in-production-please"
}
