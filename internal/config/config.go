package config

import (
	"os"
	"time"
)

type Config struct {
	Server struct {
		Port         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
	}
	Database struct {
		SupabaseURL string
		SupabaseKey string
	}
	JWT struct {
		AccessSecret  string
		RefreshSecret string
		AccessExpiry  time.Duration
		RefreshExpiry time.Duration
	}
	Environment string // "development", "production", etc.
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	cfg := &Config{}

	// Server config
	cfg.Server.Port = getEnv("SERVER_PORT", "8081")
	cfg.Server.ReadTimeout = getDurationEnv("SERVER_READ_TIMEOUT", 5*time.Second)
	cfg.Server.WriteTimeout = getDurationEnv("SERVER_WRITE_TIMEOUT", 5*time.Second)
	cfg.Server.IdleTimeout = getDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second)

	// Database config
	cfg.Database.SupabaseURL = getEnv("SUPABASE_URL", "")
	cfg.Database.SupabaseKey = getEnv("SUPABASE_ANON", "")

	// JWT config
	cfg.JWT.AccessSecret = getEnv("JWT_ACCESS_SECRET", "dev-access-secret")
	cfg.JWT.RefreshSecret = getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret")
	cfg.JWT.AccessExpiry = getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute)
	cfg.JWT.RefreshExpiry = getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour)

	// Environment
	cfg.Environment = getEnv("ENV", "development")

	return cfg
}

// Helper functions
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	str := os.Getenv(key)
	if str == "" {
		return defaultValue
	}

	val, err := time.ParseDuration(str)
	if err != nil {
		return defaultValue
	}
	return val
}
