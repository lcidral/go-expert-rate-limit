package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RedisAddr      string
	IPLimit        int
	TokenLimit     int
	IPDuration     time.Duration
	IPBlockTime    time.Duration
	TokenBlockTime time.Duration
	ServerPort     string
}

func Load() *Config {
	return &Config{
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		IPLimit:        getEnvAsInt("IP_LIMIT", 5),
		TokenLimit:     getEnvAsInt("TOKEN_LIMIT", 10),
		IPDuration:     getEnvAsDuration("IP_DURATION", "1s"),
		IPBlockTime:    getEnvAsDuration("IP_BLOCK_TIME", "5m"),
		TokenBlockTime: getEnvAsDuration("TOKEN_BLOCK_TIME", "6m"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}
