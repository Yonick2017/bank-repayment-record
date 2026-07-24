package config

import (
	"os"
	"path/filepath"
	"testing"

	"bank-repayment-record/backend/internal/auth"
)

const testPasswordHash = "e2186dbdb1bb4193608605e84f33208765b5693b55edd4f730a719a100eeea6f"
const testSessionSecret = "test-session-secret"

func validAuthYAML() string {
	return `
auth:
  password_hash: "` + testPasswordHash + `"
  session_secret: "` + testSessionSecret + `"
`
}

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
` + validAuthYAML()
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
	if cfg.Auth.PasswordHash != testPasswordHash {
		t.Fatalf("unexpected password hash %q", cfg.Auth.PasswordHash)
	}
	if cfg.Auth.SessionSecret != testSessionSecret {
		t.Fatalf("unexpected session secret %q", cfg.Auth.SessionSecret)
	}
	if cfg.Auth.SessionDays != auth.DefaultSessionDays {
		t.Fatalf("expected default session days %d, got %d", auth.DefaultSessionDays, cfg.Auth.SessionDays)
	}
	if cfg.BeianText != "" {
		t.Fatalf("expected empty beian text when omitted, got %q", cfg.BeianText)
	}

	dsn, err := cfg.MySQL.DSN()
	if err != nil {
		t.Fatalf("build dsn: %v", err)
	}
	if dsn == "" {
		t.Fatalf("expected non-empty dsn")
	}
}

func TestLoadFromPathParsesBeianText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `
mysql:
  host: "127.0.0.1"
  user: "repay"
  database: "bank_repayment"
server:
  beian_text: "  粤ICP备12345678号  "
` + validAuthYAML()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.BeianText != "粤ICP备12345678号" {
		t.Fatalf("unexpected beian text %q", cfg.BeianText)
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
` + validAuthYAML()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatalf("expected error for missing mysql.host")
	}
}

func TestLoadFromPathRequiresAuthFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `
mysql:
  host: "127.0.0.1"
  user: "repay"
  database: "bank_repayment"
server: {}
auth:
  password_hash: "not-a-hash"
  session_secret: "short"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatalf("expected error for invalid auth config")
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
` + validAuthYAML()
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
` + validAuthYAML()
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
