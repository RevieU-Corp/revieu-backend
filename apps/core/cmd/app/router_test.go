package main

import (
	"testing"

	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
)

func TestBuildRouter(t *testing.T) {
	cfg := &config.Config{}
	r := buildRouter(cfg)
	if r == nil {
		t.Fatal("expected router")
	}
}
