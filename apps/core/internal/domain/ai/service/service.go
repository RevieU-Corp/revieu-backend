package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
)

// candidateCount is the exact number of polished options returned to the client.
const candidateCount = 3

// Per-attempt sampling temperatures. The first call favors faithful polishing; if it
// returns fewer than three candidates, the retry bumps creativity to coax the model
// into producing distinct alternatives.
const (
	primaryTemperature float32 = 0.7
	retryTemperature   float32 = 1.0
)

// Sentinel errors raised by the service layer.
var (
	ErrEmptyText                = errors.New("ai: review text is empty")
	ErrInsufficientCandidates   = errors.New("ai: model returned insufficient candidates after retry")
)

// AIService orchestrates Gemini-backed review polishing. It owns no DB or storage state.
type AIService struct {
	client GeminiClient
	cfg    config.GeminiConfig
}

// NewAIService wires the service. The client may be nil only when the AI feature is
// intentionally disabled at boot — in that case any call to PolishReview will return
// an error rather than panic.
func NewAIService(client GeminiClient, cfg config.GeminiConfig) *AIService {
	return &AIService{client: client, cfg: cfg}
}

// PolishReview asks the model to rewrite the user's draft review and returns three
// polished candidates. It applies a single retry with a higher sampling temperature
// when the first call yields fewer than three usable candidates.
func (s *AIService) PolishReview(ctx context.Context, req dto.PolishRequest) (dto.PolishResponse, error) {
	if s.client == nil {
		return dto.PolishResponse{}, fmt.Errorf("%w: ai client not configured", ErrGeminiUpstream)
	}
	if req.Text == "" {
		return dto.PolishResponse{}, ErrEmptyText
	}

	candidates, err := s.client.Generate(ctx, req, primaryTemperature)
	if err != nil {
		return dto.PolishResponse{}, err
	}

	if len(candidates) < candidateCount {
		retry, retryErr := s.client.Generate(ctx, req, retryTemperature)
		if retryErr != nil {
			return dto.PolishResponse{}, retryErr
		}
		if len(retry) > len(candidates) {
			candidates = retry
		}
	}

	if len(candidates) < candidateCount {
		return dto.PolishResponse{}, ErrInsufficientCandidates
	}
	return dto.PolishResponse{Candidates: candidates[:candidateCount]}, nil
}
