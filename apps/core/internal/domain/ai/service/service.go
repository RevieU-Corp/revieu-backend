package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
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
	ErrEmptyText              = errors.New("ai: review text is empty")
	ErrInsufficientCandidates = errors.New("ai: model returned insufficient candidates after retry")
)

// AIService orchestrates Gemini-backed review polishing. It owns no DB or storage state
// of its own; the optional StyleService is consulted to personalize the prompt with the
// caller's saved writing style.
type AIService struct {
	client GeminiClient
	cfg    config.GeminiConfig
	style  *StyleService
}

// NewAIService wires the service. The client may be nil only when the AI feature is
// intentionally disabled at boot — in that case any call to PolishReview will return
// an error rather than panic.
func NewAIService(client GeminiClient, cfg config.GeminiConfig) *AIService {
	return &AIService{client: client, cfg: cfg}
}

// WithStyle attaches a StyleService for per-user prompt personalization. Returns the
// receiver to allow fluent wiring at boot. Passing a nil StyleService is fine — the
// service then behaves exactly like the un-personalized polish path.
func (s *AIService) WithStyle(style *StyleService) *AIService {
	s.style = style
	return s
}

// PolishReview asks the model to rewrite the user's draft review and returns three
// polished candidates. It applies a single retry with a higher sampling temperature
// when the first call yields fewer than three usable candidates. When UseStyle is true
// and the caller has a derived profile, the request is enriched with the profile and
// sample snippets before reaching the prompt builder.
func (s *AIService) PolishReview(ctx context.Context, req dto.PolishRequest) (dto.PolishResponse, error) {
	if s.client == nil {
		return dto.PolishResponse{}, fmt.Errorf("%w: ai client not configured", ErrGeminiUpstream)
	}
	if req.Text == "" {
		return dto.PolishResponse{}, ErrEmptyText
	}

	styleApplied := s.maybeApplyStyle(ctx, &req)

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
	return dto.PolishResponse{Candidates: candidates[:candidateCount], StyleApplied: styleApplied}, nil
}

// maybeApplyStyle looks up the user's writing-style profile and, when one exists,
// mutates req to carry the profile into prompt assembly. Returns whether style was
// actually applied so the response can advertise it. Errors are intentionally
// swallowed (logged at warn) — a style-lookup failure must never break polish.
func (s *AIService) maybeApplyStyle(ctx context.Context, req *dto.PolishRequest) bool {
	if !req.UseStyle || req.UserID <= 0 || s.style == nil {
		return false
	}
	profile, samples, err := s.style.GetStyle(ctx, req.UserID)
	if err != nil {
		logger.Warn(ctx, "ai.polish: style lookup failed; continuing without personalization",
			"user_id", req.UserID, "error", err.Error())
		return false
	}
	if profile == nil {
		return false
	}
	req.StyleProfile = profile
	req.StyleSamples = samples
	return true
}
