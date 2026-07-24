package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromPathUsesDefaultsAndParsesMySQL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `
mysql:
  host: "db.example.com"
  port: 3307
  user: "repay"
  password: "secret"
  database: "bank_repayment"

server:
  cors_allowed_origins:
    - "http://localhost:3000"
    - " https://example.com "
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.MySQL.Host != "db.example.com" {
		t.Fatalf("unexpected host %q", cfg.MySQL.Host)
	}
	if cfg.MySQL.Port != 3307 {
		t.Fatalf("unexpected port %d", cfg.MySQL.Port)
	}
	if cfg.MySQL.User != "repay" || cfg.MySQL.Password != "secret" || cfg.MySQL.Database != "bank_repayment" {
		t.Fatalf("unexpected mysql credentials: %#v", cfg.MySQL)
	}
	if cfg.MySQL.Params != defaultMySQLParams {
		t.Fatalf("expected default mysql params, got %q", cfg.MySQL.Params)
	}
	if cfg.Timezone != defaultTimezone {
		t.Fatalf("expected default timezone %q, got %q", defaultTimezone, cfg.Timezone)
	}
	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %q, got %q", defaultPort, cfg.Port)
	}
	if cfg.FrontendDistDir != defaultFrontendDistDir {
		t.Fatalf("expected default frontend dist %q, got %q", defaultFrontendDistDir, cfg.FrontendDistDir)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 cors origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "http://localhost:3000" || cfg.CORSAllowedOrigins[1] != "https://example.com" {
		t.Fatalf("unexpected cors origins: %#v", cfg.CORSAllowedOrigins)
	}

	dsn, err := cfg.MySQL.DSN()
	if err != nil {
		t.Fatalf("build dsn: %v", err)
	}
	if dsn == "" {
		t.Fatalf("expected non-empty dsn")
	}
}

func TestLoadFromPathRequiresMySQLFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `
mysql:
  host: ""
  user: "repay"
  database: "bank_repayment"
server: {}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatalf("expected error for missing mysql.host")
	}
}

func TestLoadFromPathRejectsMissingFile(t *testing.T) {
	_, err := LoadFromPath(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatalf("expected error for missing config file")
	}
}

func TestLoadFromPathRejectsInvalidTimezone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `
mysql:
  host: "127.0.0.1"
  user: "repay"
  database: "bank_repayment"
server:
  timezone: "Not/AZone"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatalf("expected error for invalid timezone")
	}
}

func TestLoadUsesCONFIG_PATH(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `
mysql:
  host: "127.0.0.1"
  user: "repay"
  password: "x"
  database: "bank_repayment"
server:
  port: "9090"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("CONFIG_PATH", path)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %q", cfg.Port)
	}
}
