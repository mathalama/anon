package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv              string
	LogLevel            string
	Port                string
	GRPCPort            string
	NATSURL             string
	RepoDriver          string
	RedisURL            string
	MatchTimeoutSec     int
	MatchFilterDropSec  int
	UserServiceURL      string
	ChatServiceURL      string
	UserServiceGRPCURL  string
	InternalToken       string
}

func Load() *Config {
	return &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		Port:               getEnv("PORT", ""),
		GRPCPort:           getEnv("GRPC_PORT", "50082"),
		NATSURL:            getEnv("NATS_URL", "nats://nats:4222"),
		RepoDriver:         getEnv("REPO_DRIVER", "redis"), // memory|redis
		RedisURL:           getEnv("REDIS_URL", ""),
		MatchTimeoutSec:    getIntEnv("MATCH_TIMEOUT_SEC", 60),
		MatchFilterDropSec: getIntEnv("MATCH_FILTER_DROP_SEC", 30),
		UserServiceURL:     getEnv("USER_SERVICE_URL", ""),
		ChatServiceURL:     getEnv("CHAT_SERVICE_URL", ""),
		UserServiceGRPCURL: getEnv("USER_SERVICE_GRPC_URL", "user-service:50081"),
		InternalToken:      getEnv("INTERNAL_TOKEN", ""),
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
