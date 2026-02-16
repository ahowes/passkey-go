package config

import (
	"os"
	"strings"
)

type Config struct {
	ListenAddr       string
	DatabaseDSN      string
	DatabaseUser     string
	DatabasePassword string
	RPDisplayName    string
	RPID             string
	RPOrigins        []string
	SessionSecret    string
}

func Load() *Config {
	cfg := &Config{
		ListenAddr:       envOrDefault("LISTEN_ADDR", ":8080"),
		DatabaseDSN:      envOrDefault("DATABASE_DSN", "postgres://localhost:5432/test_db?sslmode=disable"),
		DatabaseUser:     envOrDefault("DATABASE_USER", "admin"),
		DatabasePassword: envOrDefault("DATABASE_PASSWORD", "root"),
		RPDisplayName:    envOrDefault("WEBAUTHN_RP_DISPLAY_NAME", "Passkey Go"),
		RPID:             envOrDefault("WEBAUTHN_RP_ID", "localhost"),
		SessionSecret:    envOrDefault("SESSION_SECRET", "super-secret-key-change-me-in-prod"),
	}

	origins := envOrDefault("WEBAUTHN_RP_ORIGINS", "http://localhost:8080")
	cfg.RPOrigins = strings.Split(origins, ",")

	return cfg
}

func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
