package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
		if r.Method != http.MethodPost || r.URL.Path != "/accounts/account/storage/kv/namespaces/namespace/values/bulk/get" {
			t.Errorf("expected bulk/get POST, got %s %s", r.Method, r.URL.Path)
		}
		var req struct{ Keys []string `json:"keys"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { t.Error(err); return }
		values := map[string]string{}
		for _, k := range req.Keys {
			if v, ok := map[string]string{
				"DB_HOST":  "remote-db",
				"local/DB_HOST": "remote-db",
			}[k]; ok { values[k] = v }
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": map[string]interface{}{"values": values}, "success": true})
	}))
	defer server.Close()
	p := &cloudflareKVProvider{client: server.Client(), accountID: "account", namespaceID: "namespace", token: "token", prefix: "local/", baseURL: server.URL}
	values, err := p.Load(context.Background(), []string{"DB_HOST"})
	if err != nil { t.Fatal(err) }
	if values["DB_HOST"] != "remote-db" { t.Fatalf("got %q", values["DB_HOST"]) }
}

func TestCloudflareKVUnprefixedFallback(t *testing.T) {
	// Prefixed key absent -> falls back to unprefixed key.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Keys []string `json:"keys"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { t.Error(err); return }
		values := map[string]string{}
		for _, k := range req.Keys {
			if k == "DB_HOST" { values[k] = "fallback-host" } // only unprefixed exists
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": map[string]interface{}{"values": values}, "success": true})
	}))
	defer server.Close()
	p := &cloudflareKVProvider{client: server.Client(), accountID: "account", namespaceID: "namespace", token: "token", prefix: "cluster/", baseURL: server.URL}
	values, err := p.Load(context.Background(), []string{"DB_HOST"})
	if err != nil { t.Fatal(err) }
	if values["DB_HOST"] != "fallback-host" { t.Fatalf("got %q", values["DB_HOST"]) }
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	_ = os.Unsetenv("JWT_SECRET")
	if _, err := decode(map[string]string{}); err == nil { t.Fatal("expected JWT_SECRET validation error") }
}
