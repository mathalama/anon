package config

import (
	"os"
	"strings"
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
	AllowedOrigins        []string
	DevAllowedOrigins     []string
}

func Load() *Config {
	return &Config{
		AppEnv:                getEnv("APP_ENV", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		Port:                  getEnv("PORT", ""),
		RedisURL:              getEnv("REDIS_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		UserServiceURL:        getEnv("USER_SERVICE_URL", ""),
		MatchmakingServiceURL: getEnv("MATCHMAKING_SERVICE_URL", ""),
		ChatServiceURL:        getEnv("CHAT_SERVICE_URL", ""),
		ModerationServiceURL:  getEnv("MODERATION_SERVICE_URL", ""),
		AllowedOrigins:        strings.Split(getEnv("ALLOWED_ORIGINS", ""), ","),
		DevAllowedOrigins:     strings.Split(getEnv("DEV_ALLOWED_ORIGINS", ""), ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
