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

	// VOD (contracts/vod-env.md) — local por default; S3 só na demo AWS (Fase 3).
	VODBackend     string
	VODLocalDir    string
	VODS3Bucket    string
	VODPlaybackTTL time.Duration
	AWSRegion      string

	// Live / IVS (contracts/live-env.md) — stub por default; ivs só na demo AWS (Fase 3).
	LiveBackend       string
	IVSIngestEndpoint string
	IVSStreamKey      string
	IVSPlaybackURL    string
	IVSChannelARN     string
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

		VODBackend:     getEnv("VOD_BACKEND", "local"),
		VODLocalDir:    getEnv("VOD_LOCAL_DIR", "./data/vod"),
		VODS3Bucket:    getEnv("VOD_S3_BUCKET", ""),
		VODPlaybackTTL: getDurationEnv("VOD_PLAYBACK_TTL", 15*time.Minute),
		AWSRegion:      getEnv("AWS_REGION", ""),

		LiveBackend:       normalizeLiveBackend(getEnv("LIVE_BACKEND", "stub")),
		IVSIngestEndpoint: getEnv("IVS_INGEST_ENDPOINT", ""),
		IVSStreamKey:      getEnv("IVS_STREAM_KEY", ""),
		IVSPlaybackURL:    getEnv("IVS_PLAYBACK_URL", ""),
		IVSChannelARN:     getEnv("IVS_CHANNEL_ARN", ""),
	}
}

// LiveBackendValid reporta se LIVE_BACKEND é stub ou ivs (após normalização).
func LiveBackendValid(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "stub", "ivs":
		return true
	default:
		return false
	}
}

func normalizeLiveBackend(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

// IVSConfigured indica se ingest, stream key e playback estão preenchidos.
func (c Config) IVSConfigured() bool {
	return strings.TrimSpace(c.IVSIngestEndpoint) != "" &&
		strings.TrimSpace(c.IVSStreamKey) != "" &&
		strings.TrimSpace(c.IVSPlaybackURL) != ""
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
