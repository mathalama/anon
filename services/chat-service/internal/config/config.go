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
	RepoDriver           string
	RedisURL             string
	WSMaxConn            int
	MsgRetentionDays     int
	ModerationServiceURL string
	InternalToken        string
	AllowedOrigins       []string
}

func Load() *Config {
	return &Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		Port:                 getEnv("PORT", ""),
		DBURL:                getEnv("DB_URL", ""),
		RepoDriver:           getEnv("REPO_DRIVER", "postgres"), // memory|postgres
		RedisURL:             getEnv("REDIS_URL", ""),
		WSMaxConn:            getIntEnv("WS_MAX_CONN", 1000),
		MsgRetentionDays:     getIntEnv("MSG_RETENTION_DAYS", 7),
		ModerationServiceURL: getEnv("MODERATION_SERVICE_URL", ""),
		InternalToken:        getEnv("INTERNAL_TOKEN", ""),
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
