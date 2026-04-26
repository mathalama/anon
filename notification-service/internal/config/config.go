package config

import (
	"os"
)

type Config struct {
	AppEnv           string
	LogLevel         string
	Port             string
	InternalToken    string
	TelegramBotToken string
	SMTPHost         string
	SMTPPort         string
}

func Load() *Config {
	return &Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "debug"),
		Port:             getEnv("PORT", "8085"),
		InternalToken:    getEnv("INTERNAL_TOKEN", "dev-internal-token"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnv("SMTP_PORT", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
