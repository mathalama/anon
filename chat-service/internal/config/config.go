package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv               string
	LogLevel             string
	Port                 string
	DBURL                string
	RedisURL             string
	WSMaxConn            int
	MsgRetentionDays     int
	ModerationServiceURL string
}

func Load() *Config {
	return &Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		LogLevel:             getEnv("LOG_LEVEL", "debug"),
		Port:                 getEnv("PORT", "8083"),
		DBURL:                getEnv("DB_URL", "postgres://user:pass@localhost:5432/chat_db?sslmode=disable"),
		RedisURL:             getEnv("REDIS_URL", "redis:6379"),
		WSMaxConn:            getIntEnv("WS_MAX_CONN", 10000),
		MsgRetentionDays:     getIntEnv("MSG_RETENTION_DAYS", 30),
		ModerationServiceURL: getEnv("MODERATION_SERVICE_URL", "http://moderation-service:8084"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
