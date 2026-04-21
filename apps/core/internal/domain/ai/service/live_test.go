package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
)

// TestPolishReview_LiveSmoke hits the real Gemini API when GEMINI_API_KEY is set.
// It is skipped in CI (no key) and intentionally prints the candidates via t.Log
// so you can eyeball the output. Invoke with:
//
//	GEMINI_API_KEY=... go test -run TestPolishReview_LiveSmoke -v ./internal/domain/ai/service
func TestPolishReview_LiveSmoke(t *testing.T) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		t.Skip("GEMINI_API_KEY not set; skipping live smoke test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cfg := config.GeminiConfig{
		Backend:        os.Getenv("GEMINI_BACKEND"),
		APIKey:         key,
		Model:          envOr("GEMINI_MODEL", "gemini-2.5-flash"),
		TimeoutSeconds: 30,
	}
	client, err := NewGeminiClient(ctx, cfg)
	if err != nil {
		t.Fatalf("NewGeminiClient: %v", err)
	}

	svc := NewAIService(client, cfg)

	req := dto.PolishRequest{
		Text:             "coffee was amazing and the staff very friendly but tables too close together",
		MerchantName:     "Blue Bottle Cafe",
		BusinessCategory: "cafe",
		Language:         "en",
		Rating: dto.PolishRating{
			Overall:     ptrFloat(4.0),
			Service:     ptrFloat(5.0),
			Environment: ptrFloat(3.0),
		},
	}

	resp, err := svc.PolishReview(ctx, req)
	if err != nil {
		t.Fatalf("PolishReview: %v", err)
	}

	if len(resp.Candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d: %v", len(resp.Candidates), resp.Candidates)
	}

	t.Log("--- polished candidates ---")
	for i, c := range resp.Candidates {
		t.Logf("[%d] %s", i+1, c)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
