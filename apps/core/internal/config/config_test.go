package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  address: ":8080"
  port: 8080
  mode: "debug"

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  database: "testdb"
  username: "testuser"
  password: "testpass"

logger:
  level: "info"
  format: "json"

jwt:
  secret: "test-secret"
  expire_hour: 24

oauth:
  google:
    client_id: "test-client-id"
    client_secret: "test-client-secret"

frontend_url: "http://localhost:3000"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// Set CONFIG_PATH env var
	os.Setenv("CONFIG_PATH", configPath)
	defer os.Unsetenv("CONFIG_PATH")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify values
	if cfg.Server.Address != ":8080" {
		t.Errorf("Server.Address = %v, want :8080", cfg.Server.Address)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %v, want localhost", cfg.Database.Host)
	}
	if cfg.JWT.Secret != "test-secret" {
		t.Errorf("JWT.Secret = %v, want test-secret", cfg.JWT.Secret)
	}
	if cfg.OAuth.Google.ClientID != "test-client-id" {
		t.Errorf("OAuth.Google.ClientID = %v, want test-client-id", cfg.OAuth.Google.ClientID)
	}
	if cfg.FrontendURL != "http://localhost:3000" {
		t.Errorf("FrontendURL = %v, want http://localhost:3000", cfg.FrontendURL)
	}
}

func TestLoad_EnvVarExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  address: ":8080"
  port: 8080
  mode: "debug"

database:
  driver: "postgres"
  host: "${DB_HOST}"
  port: 5432
  database: "testdb"
  username: "testuser"
  password: "testpass"

logger:
  level: "info"
  format: "json"

jwt:
  secret: "${JWT_SECRET}"
  expire_hour: 24

oauth:
  google:
    client_id: "${GOOGLE_CLIENT_ID}"
    client_secret: "${GOOGLE_CLIENT_SECRET}"

frontend_url: "${FRONTEND_URL}"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// Set environment variables
	os.Setenv("CONFIG_PATH", configPath)
	os.Setenv("DB_HOST", "10.0.0.1")
	os.Setenv("JWT_SECRET", "env-jwt-secret")
	os.Setenv("GOOGLE_CLIENT_ID", "env-google-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "env-google-client-secret")
	os.Setenv("FRONTEND_URL", "https://example.com")
	defer func() {
		os.Unsetenv("CONFIG_PATH")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("GOOGLE_CLIENT_ID")
		os.Unsetenv("GOOGLE_CLIENT_SECRET")
		os.Unsetenv("FRONTEND_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Host != "10.0.0.1" {
		t.Errorf("Database.Host = %v, want 10.0.0.1", cfg.Database.Host)
	}
	if cfg.JWT.Secret != "env-jwt-secret" {
		t.Errorf("JWT.Secret = %v, want env-jwt-secret", cfg.JWT.Secret)
	}
	if cfg.OAuth.Google.ClientID != "env-google-client-id" {
		t.Errorf("OAuth.Google.ClientID = %v, want env-google-client-id", cfg.OAuth.Google.ClientID)
	}
	if cfg.OAuth.Google.ClientSecret != "env-google-client-secret" {
		t.Errorf("OAuth.Google.ClientSecret = %v, want env-google-client-secret", cfg.OAuth.Google.ClientSecret)
	}
	if cfg.FrontendURL != "https://example.com" {
		t.Errorf("FrontendURL = %v, want https://example.com", cfg.FrontendURL)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	os.Setenv("CONFIG_PATH", "/nonexistent/path/config.yaml")
	defer os.Unsetenv("CONFIG_PATH")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidContent := `
server:
  address: ":8080"
  port: [invalid yaml
`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	os.Setenv("CONFIG_PATH", configPath)
	defer os.Unsetenv("CONFIG_PATH")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}
