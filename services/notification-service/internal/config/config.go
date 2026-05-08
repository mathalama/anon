package config

import (
	"os"
)

type Config struct {
	AppEnv           string
	LogLevel         string
	Port             string
	InternalToken    string
	SMTPHost         string
	SMTPPort         string
	NATSURL          string
}

func Load() *Config {
	return &Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Port:             getEnv("PORT", ""),
		InternalToken:    getEnv("INTERNAL_TOKEN", ""),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnv("SMTP_PORT", ""),
		NATSURL:          getEnv("NATS_URL", "nats://nats:4222"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
