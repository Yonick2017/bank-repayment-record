package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultDBPath            = "data/repayments.db"
	defaultTimezone          = "Asia/Shanghai"
	defaultPort              = "8080"
	defaultCORSAllowedOrigin = "http://localhost:5173,http://127.0.0.1:5173"
	defaultFrontendDistDir   = "../frontend/dist"
)

type Config struct {
	DBPath             string
	Timezone           string
	Port               string
	CORSAllowedOrigins []string
	FrontendDistDir    string
}

func Load() (Config, error) {
	cfg := Config{
		DBPath:             envOrDefault("DB_PATH", defaultDBPath),
		Timezone:           envOrDefault("TIMEZONE", defaultTimezone),
		Port:               envOrDefault("PORT", defaultPort),
		CORSAllowedOrigins: parseCommaSeparatedValues(envOrDefault("CORS_ALLOWED_ORIGINS", defaultCORSAllowedOrigin)),
		FrontendDistDir:    envOrDefault("FRONTEND_DIST_DIR", defaultFrontendDistDir),
	}

	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		return Config{}, fmt.Errorf("invalid TIMEZONE %q: %w", cfg.Timezone, err)
	}
	if err := ValidateDBPathWritable(cfg.DBPath); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ValidateDBPathWritable(path string) error {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "." || cleanPath == "" {
		return fmt.Errorf("DB_PATH must point to a database file")
	}

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create DB_PATH parent directory %q: %w", dir, err)
	}

	file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("cannot write DB_PATH %q: %w", cleanPath, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("cannot close DB_PATH %q during startup check: %w", cleanPath, err)
	}

	return nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseCommaSeparatedValues(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	if len(values) == 0 {
		return parseCommaSeparatedValues(defaultCORSAllowedOrigin)
	}
	return values
}
