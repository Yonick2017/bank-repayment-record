package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bank-repayment-record/backend/internal/auth"

	"github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

const (
	defaultTimezone        = "Asia/Shanghai"
	defaultPort            = "8080"
	defaultFrontendDistDir = "../frontend/dist"
	defaultMySQLParams     = "parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci"
)

var defaultCORSAllowedOrigins = []string{
	"http://localhost:5173",
	"http://127.0.0.1:5173",
}

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Params   string
}

type Config struct {
	MySQL              MySQLConfig
	Auth               auth.Config
	Timezone           string
	Port               string
	CORSAllowedOrigins []string
	FrontendDistDir    string
}

type fileConfig struct {
	MySQL  fileMySQLConfig  `yaml:"mysql"`
	Server fileServerConfig `yaml:"server"`
	Auth   fileAuthConfig   `yaml:"auth"`
}

type fileMySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Params   string `yaml:"params"`
}

type fileServerConfig struct {
	Port               string   `yaml:"port"`
	Timezone           string   `yaml:"timezone"`
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins"`
	FrontendDistDir    string   `yaml:"frontend_dist_dir"`
}

type fileAuthConfig struct {
	PasswordHash  string `yaml:"password_hash"`
	SessionSecret string `yaml:"session_secret"`
	SessionDays   int    `yaml:"session_days"`
}

func Load() (Config, error) {
	path, err := resolveConfigPath()
	if err != nil {
		return Config{}, err
	}
	return LoadFromPath(path)
}

func LoadFromPath(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var file fileConfig
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	sessionDays := file.Auth.SessionDays
	if sessionDays <= 0 {
		sessionDays = auth.DefaultSessionDays
	}

	cfg := Config{
		MySQL: MySQLConfig{
			Host:     strings.TrimSpace(file.MySQL.Host),
			Port:     file.MySQL.Port,
			User:     strings.TrimSpace(file.MySQL.User),
			Password: file.MySQL.Password,
			Database: strings.TrimSpace(file.MySQL.Database),
			Params:   strings.TrimSpace(file.MySQL.Params),
		},
		Auth: auth.Config{
			PasswordHash:  strings.ToLower(strings.TrimSpace(file.Auth.PasswordHash)),
			SessionSecret: strings.TrimSpace(file.Auth.SessionSecret),
			SessionDays:   sessionDays,
		},
		Timezone:           strings.TrimSpace(file.Server.Timezone),
		Port:               strings.TrimSpace(file.Server.Port),
		CORSAllowedOrigins: normalizeOrigins(file.Server.CORSAllowedOrigins),
		FrontendDistDir:    strings.TrimSpace(file.Server.FrontendDistDir),
	}

	if cfg.MySQL.Port == 0 {
		cfg.MySQL.Port = 3306
	}
	if cfg.MySQL.Params == "" {
		cfg.MySQL.Params = defaultMySQLParams
	}
	if cfg.Timezone == "" {
		cfg.Timezone = defaultTimezone
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.FrontendDistDir == "" {
		cfg.FrontendDistDir = defaultFrontendDistDir
	}
	if len(cfg.CORSAllowedOrigins) == 0 {
		cfg.CORSAllowedOrigins = append([]string{}, defaultCORSAllowedOrigins...)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.MySQL.Host == "" {
		return fmt.Errorf("mysql.host is required")
	}
	if c.MySQL.User == "" {
		return fmt.Errorf("mysql.user is required")
	}
	if c.MySQL.Database == "" {
		return fmt.Errorf("mysql.database is required")
	}
	if c.MySQL.Port <= 0 || c.MySQL.Port > 65535 {
		return fmt.Errorf("mysql.port must be between 1 and 65535")
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("invalid server.timezone %q: %w", c.Timezone, err)
	}
	if err := auth.ValidatePasswordHash(c.Auth.PasswordHash); err != nil {
		return err
	}
	if err := auth.ValidateSessionSecret(c.Auth.SessionSecret); err != nil {
		return err
	}
	if c.Auth.SessionDays <= 0 {
		return fmt.Errorf("auth.session_days must be positive")
	}
	return nil
}

func (m MySQLConfig) DSN() (string, error) {
	cfg := mysql.NewConfig()
	cfg.User = m.User
	cfg.Passwd = m.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", m.Host, m.Port)
	cfg.DBName = m.Database
	cfg.Params = map[string]string{}
	cfg.ParseTime = true

	params := strings.TrimPrefix(strings.TrimSpace(m.Params), "?")
	if params != "" {
		values, err := parseQueryParams(params)
		if err != nil {
			return "", fmt.Errorf("mysql.params: %w", err)
		}
		for key, value := range values {
			switch key {
			case "parseTime":
				cfg.ParseTime = strings.EqualFold(value, "true")
			case "loc":
				loc, err := time.LoadLocation(value)
				if err != nil {
					return "", fmt.Errorf("invalid mysql.params loc %q: %w", value, err)
				}
				cfg.Loc = loc
			case "charset":
				cfg.Params["charset"] = value
			case "collation":
				cfg.Collation = value
			default:
				cfg.Params[key] = value
			}
		}
	}
	if cfg.Loc == nil {
		cfg.Loc = time.Local
	}
	return cfg.FormatDSN(), nil
}

func parseQueryParams(raw string) (map[string]string, error) {
	parts := strings.Split(raw, "&")
	values := make(map[string]string, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid param %q", part)
		}
		values[key] = value
	}
	return values, nil
}

func resolveConfigPath() (string, error) {
	if explicit := strings.TrimSpace(os.Getenv("CONFIG_PATH")); explicit != "" {
		return explicit, nil
	}

	candidates := []string{
		"config.yaml",
		filepath.Join("backend", "config.yaml"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("config file not found; set CONFIG_PATH or create config.yaml (see config.example.yaml)")
}

func normalizeOrigins(origins []string) []string {
	values := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	return values
}
