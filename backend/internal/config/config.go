package config

import (
	"os"
	"time"
)

type Config struct {
	Port          string
	DatabaseURL   string
	CORSOrigin    string
	JWTSecret     string
	JWTExpiration time.Duration
	AdminEmail    string
	AdminPassword string
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://estudaja:estudaja@localhost:5432/estudaja?sslmode=disable"),
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:5173"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@estudaja.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}
