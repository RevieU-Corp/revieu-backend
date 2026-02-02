package main

import (
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
)

func TestBuildRouter(t *testing.T) {
	cfg := &config.Config{}
	r := buildRouter(cfg)
	if r == nil {
		t.Fatal("expected router")
	}
}
