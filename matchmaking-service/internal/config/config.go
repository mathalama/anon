package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv              string
	LogLevel            string
	Port                string
	RepoDriver          string
	RedisURL            string
	MatchTimeoutSec     int
	MatchFilterDropSec  int
	UserServiceURL      string
	ChatServiceURL      string
	InternalToken       string
}

func Load() *Config {
	return &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "debug"),
		Port:               getEnv("PORT", "8082"),
		RepoDriver:         getEnv("REPO_DRIVER", "memory"), // memory|redis
		RedisURL:           getEnv("REDIS_URL", "redis:6379"),
		MatchTimeoutSec:    getIntEnv("MATCH_TIMEOUT_SEC", 60),
		MatchFilterDropSec: getIntEnv("MATCH_FILTER_DROP_SEC", 30),
		UserServiceURL:     getEnv("USER_SERVICE_URL", "http://user-service:8081"),
		ChatServiceURL:     getEnv("CHAT_SERVICE_URL", "http://chat-service:8083"),
		InternalToken:      getEnv("INTERNAL_TOKEN", "dev-internal-token"),
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
