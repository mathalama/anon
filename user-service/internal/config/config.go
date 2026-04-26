package config

import (
	"os"
	"time"
)

type Config struct {
	AppEnv         string
	LogLevel       string
	DBURL          string
	RepoDriver     string
	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	InternalToken      string
	TelegramBotToken   string
	Port               string
}

func Load() *Config {
	return &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "debug"),
		DBURL:          getEnv("DB_URL", "postgres://user:pass@localhost:5432/users_db?sslmode=disable"),
		RepoDriver:     getEnv("REPO_DRIVER", "memory"), // memory|postgres
		JWTSecret:      getEnv("JWT_SECRET", "very-secret-key"),
		JWTAccessTTL:   getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:  getDurationEnv("JWT_REFRESH_TTL", 168*time.Hour),
		InternalToken:    getEnv("INTERNAL_TOKEN", "dev-internal-token"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		Port:             getEnv("PORT", "8081"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
