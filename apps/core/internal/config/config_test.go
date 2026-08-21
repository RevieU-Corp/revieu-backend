package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DB_HOST", "db.example")
	cfg, err := Load()
	if err != nil { t.Fatal(err) }
	if cfg.Server.Port != 9090 || cfg.Database.Host != "db.example" || cfg.JWT.Secret != "env-secret" { t.Fatalf("unexpected config: %+v", cfg) }
}

func TestCloudflareKVOverridesEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" { t.Error("missing authorization") }
		if r.Method != http.MethodGet { t.Errorf("expected GET, got %s", r.Method) }
		if strings.HasSuffix(r.URL.Path, "/values/DB_HOST") { _, _ = w.Write([]byte("remote-db")); return }
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	p := &cloudflareKVProvider{client: server.Client(), accountID: "account", namespaceID: "namespace", token: "token", baseURL: server.URL}
	values, err := p.Load(context.Background(), []string{"DB_HOST"})
	if err != nil { t.Fatal(err) }
	if values["DB_HOST"] != "remote-db" { t.Fatalf("got %q", values["DB_HOST"]) }
}

func TestEnvironmentOverridesCloudflareKV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/values/DB_HOST") { _, _ = w.Write([]byte("remote-db")); return }
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	t.Setenv("DB_HOST", "env-db")
	t.Setenv("JWT_SECRET", "env-secret")
	p := &cloudflareKVProvider{client: server.Client(), accountID: "account", namespaceID: "namespace", token: "token", baseURL: server.URL}
	values, err := p.Load(context.Background(), []string{"DB_HOST"})
	if err != nil { t.Fatal(err) }
	if values["DB_HOST"] != "remote-db" { t.Fatalf("unexpected remote value %q", values["DB_HOST"]) }
	if os.Getenv("DB_HOST") != "env-db" { t.Fatal("test environment changed") }
}


func TestLoadRequiresJWTSecret(t *testing.T) {
	_ = os.Unsetenv("JWT_SECRET")
	if _, err := decode(map[string]string{}); err == nil { t.Fatal("expected JWT_SECRET validation error") }
}
