package config

import (
	"os"
)

type Config struct {
	AppEnv                 string
	LogLevel               string
	Port                   string
	DBURL                  string
	UserServiceURL         string
	ChatServiceURL         string
	NotificationServiceURL string
}

func Load() *Config {
	return &Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		LogLevel:               getEnv("LOG_LEVEL", "debug"),
		Port:                   getEnv("PORT", "8084"),
		DBURL:                  getEnv("DB_URL", "postgres://user:pass@localhost:5432/moderation_db?sslmode=disable"),
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://user-service:8081"),
		ChatServiceURL:         getEnv("CHAT_SERVICE_URL", "http://chat-service:8083"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8085"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
