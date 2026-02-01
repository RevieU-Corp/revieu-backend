package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig   `yaml:"server"`
	Database    DatabaseConfig `yaml:"database"`
	Logger      LoggerConfig   `yaml:"logger"`
	JWT         JWTConfig      `yaml:"jwt"`
	OAuth       OAuthConfig    `yaml:"oauth"`
	SMTP        SMTPConfig     `yaml:"smtp"`
	FrontendURL string         `yaml:"frontend_url"`
}

// SMTPConfig holds SMTP email configuration
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	UseTLS   bool   `yaml:"use_tls"`
}

// OAuthConfig holds OAuth provider configurations
type OAuthConfig struct {
	Google GoogleOAuthConfig `yaml:"google"`
}

// GoogleOAuthConfig holds Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireHour int    `yaml:"expire_hour"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address     string `yaml:"address"`
	Port        int    `yaml:"port"`
	Mode        string `yaml:"mode"`          // debug, release, test
	APIBasePath string `yaml:"api_base_path"` // API version prefix (e.g., /api/v1)
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

// Load reads configuration from file
func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand environment variables in JWT secret
	if strings.HasPrefix(cfg.JWT.Secret, "${") && strings.HasSuffix(cfg.JWT.Secret, "}") {
		envVar := cfg.JWT.Secret[2 : len(cfg.JWT.Secret)-1]
		cfg.JWT.Secret = os.Getenv(envVar)
	}

	// Expand environment variables in database config
	if strings.HasPrefix(cfg.Database.Password, "${") && strings.HasSuffix(cfg.Database.Password, "}") {
		envVar := cfg.Database.Password[2 : len(cfg.Database.Password)-1]
		cfg.Database.Password = os.Getenv(envVar)
	}

	// Expand environment variables in OAuth config
	if strings.HasPrefix(cfg.OAuth.Google.ClientID, "${") && strings.HasSuffix(cfg.OAuth.Google.ClientID, "}") {
		envVar := cfg.OAuth.Google.ClientID[2 : len(cfg.OAuth.Google.ClientID)-1]
		cfg.OAuth.Google.ClientID = os.Getenv(envVar)
	}
	if strings.HasPrefix(cfg.OAuth.Google.ClientSecret, "${") && strings.HasSuffix(cfg.OAuth.Google.ClientSecret, "}") {
		envVar := cfg.OAuth.Google.ClientSecret[2 : len(cfg.OAuth.Google.ClientSecret)-1]
		cfg.OAuth.Google.ClientSecret = os.Getenv(envVar)
	}
	if strings.HasPrefix(cfg.FrontendURL, "${") && strings.HasSuffix(cfg.FrontendURL, "}") {
		envVar := cfg.FrontendURL[2 : len(cfg.FrontendURL)-1]
		cfg.FrontendURL = os.Getenv(envVar)
	}

	// Expand environment variables in SMTP config
	if strings.HasPrefix(cfg.SMTP.Username, "${") && strings.HasSuffix(cfg.SMTP.Username, "}") {
		envVar := cfg.SMTP.Username[2 : len(cfg.SMTP.Username)-1]
		cfg.SMTP.Username = os.Getenv(envVar)
	}
	if strings.HasPrefix(cfg.SMTP.Password, "${") && strings.HasSuffix(cfg.SMTP.Password, "}") {
		envVar := cfg.SMTP.Password[2 : len(cfg.SMTP.Password)-1]
		cfg.SMTP.Password = os.Getenv(envVar)
	}

	return &cfg, nil
}
