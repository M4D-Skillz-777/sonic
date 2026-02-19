package config

import "os"

type Config struct {
	RedisAddr string
	Port      string
}

func Load() *Config {
	return &Config{
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		Port:      getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
