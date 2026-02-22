package token

import (
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
)

func TestHashTokenDeterministic(t *testing.T) {
	h1 := HashToken("abc123")
	h2 := HashToken("abc123")
	if h1 == "" {
		t.Fatal("expected hash to be non-empty")
	}
	if h1 != h2 {
		t.Fatalf("expected deterministic hash, got %q and %q", h1, h2)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	svc := New(config.JWTConfig{Secret: "secret", ExpireHour: 24, RefreshExpireHour: 168})

	plain, hash, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if plain == "" {
		t.Fatal("expected refresh token")
	}
	if hash == "" {
		t.Fatal("expected refresh token hash")
	}
	if hash != HashToken(plain) {
		t.Fatal("expected hash to match token")
	}
}
