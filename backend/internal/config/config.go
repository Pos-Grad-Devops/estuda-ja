package config

import (
	"os"
	"strings"
	"time"
)

// Config carrega o runtime da API a partir de env (contrato specs/.../contracts/api-env.md).
// Defaults abaixo são só para desenvolvimento local — nunca secrets de conta AWS.
type Config struct {
	Port          string
	DatabaseURL   string
	CORSOrigin    string
	JWTSecret     string
	JWTExpiration time.Duration
	AdminEmail    string
	AdminPassword string

	DemoProfessorEmail    string
	DemoProfessorPassword string
	DemoAlunoEmail        string
	DemoAlunoPassword     string
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

		DemoProfessorEmail:    getEnv("DEMO_PROFESSOR_EMAIL", "professor@estudaja.com"),
		DemoProfessorPassword: getEnv("DEMO_PROFESSOR_PASSWORD", "professor123"),
		DemoAlunoEmail:        getEnv("DEMO_ALUNO_EMAIL", "aluno@estudaja.com"),
		DemoAlunoPassword:     getEnv("DEMO_ALUNO_PASSWORD", "aluno123"),
	}
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}
