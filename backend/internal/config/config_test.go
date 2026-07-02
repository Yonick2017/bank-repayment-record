package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("DB_PATH", "")
	t.Setenv("TIMEZONE", "")
	t.Setenv("PORT", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("FRONTEND_DIST_DIR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.DBPath != defaultDBPath {
		t.Fatalf("expected default DB_PATH %q, got %q", defaultDBPath, cfg.DBPath)
	}
	if cfg.Timezone != defaultTimezone {
		t.Fatalf("expected default TIMEZONE %q, got %q", defaultTimezone, cfg.Timezone)
	}
	if cfg.Port != defaultPort {
		t.Fatalf("expected default PORT %q, got %q", defaultPort, cfg.Port)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 default CORS origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "http://localhost:5173" {
		t.Fatalf("unexpected default CORS origin %q", cfg.CORSAllowedOrigins[0])
	}
	if cfg.FrontendDistDir != defaultFrontendDistDir {
		t.Fatalf("expected default FRONTEND_DIST_DIR %q, got %q", defaultFrontendDistDir, cfg.FrontendDistDir)
	}
}

func TestValidateDBPathWritable(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "nested", "repayments.db")
	if err := ValidateDBPathWritable(dbPath); err != nil {
		t.Fatalf("expected db path to be writable: %v", err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected db file to be created: %v", err)
	}
}

func TestValidateDBPathWritableRejectsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	err := ValidateDBPathWritable(tempDir)
	if err == nil {
		t.Fatalf("expected error when DB_PATH points to a directory")
	}
}

func TestLoadParsesCORSAllowedOrigins(t *testing.T) {
	t.Setenv("DB_PATH", filepath.Join(t.TempDir(), "repayments.db"))
	t.Setenv("TIMEZONE", "Asia/Shanghai")
	t.Setenv("PORT", "8080")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://example.com ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 configured CORS origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "http://localhost:3000" || cfg.CORSAllowedOrigins[1] != "https://example.com" {
		t.Fatalf("unexpected CORS origins: %#v", cfg.CORSAllowedOrigins)
	}
}
