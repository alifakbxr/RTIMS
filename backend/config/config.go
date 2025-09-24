package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment  string
	Port         string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	RefreshSecret string
	EmailAPIKey  string
	EmailFrom    string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	AllowedOrigins []string
	RateLimit    int
}

func Load() *Config {
	return &Config{
		Environment:    getEnv("ENVIRONMENT", "development"),
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/rtims?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		RefreshSecret:  getEnv("REFRESH_SECRET", "your-super-secret-refresh-key-change-this-in-production"),
		EmailAPIKey:    getEnv("EMAIL_API_KEY", ""),
		EmailFrom:      getEnv("EMAIL_FROM", "noreply@rtims.com"),
		SMTPHost:       getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:       getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:3001"},
		RateLimit:      getEnvAsInt("RATE_LIMIT", 100),
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
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}