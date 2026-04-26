package config

import (
	"os"
)

type Config struct {
	AppEnv                string
	LogLevel              string
	Port                  string
	RedisURL              string
	JWTSecret             string
	UserServiceURL        string
	MatchmakingServiceURL string
	ChatServiceURL        string
	ModerationServiceURL  string
}

func Load() *Config {
	return &Config{
		AppEnv:                getEnv("APP_ENV", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "debug"),
		Port:                  getEnv("PORT", "8080"),
		RedisURL:              getEnv("REDIS_URL", "redis:6379"),
		JWTSecret:             getEnv("JWT_SECRET", "very-secret-key"),
		UserServiceURL:        getEnv("USER_SERVICE_URL", "http://user-service:8081"),
		MatchmakingServiceURL: getEnv("MATCHMAKING_SERVICE_URL", "http://matchmaking-service:8082"),
		ChatServiceURL:        getEnv("CHAT_SERVICE_URL", "http://chat-service:8083"),
		ModerationServiceURL:  getEnv("MODERATION_SERVICE_URL", "http://moderation-service:8084"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
