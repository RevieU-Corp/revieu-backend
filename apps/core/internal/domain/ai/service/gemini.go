package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
	"google.golang.org/genai"
)

// GeminiClient is the narrow surface the AI service depends on. It is an interface so the
// service can be unit-tested with a fake instead of the real SDK.
//
// Generate sends one request to the model with the given temperature and returns the
// candidate strings parsed out of the JSON response. The implementation is responsible
// for trimming whitespace and dropping empty entries; the caller decides whether the
// count is acceptable and whether to retry.
type GeminiClient interface {
	Generate(ctx context.Context, req dto.PolishRequest, temperature float32) ([]string, error)
}

// Sentinel errors mapped to HTTP statuses by the handler.
var (
	ErrGeminiSafetyBlocked      = errors.New("ai: content blocked by safety filter")
	ErrGeminiUpstream           = errors.New("ai: upstream model error")
	ErrGeminiRateLimited        = errors.New("ai: upstream rate limited")
	ErrGeminiInvalidResponse    = errors.New("ai: invalid response from model")
	ErrGeminiInsufficientOutput = errors.New("ai: model returned fewer candidates than requested")
)

// SafetyBlockError carries the upstream block reason for diagnostic messages.
type SafetyBlockError struct {
	Reason string
}

func (e *SafetyBlockError) Error() string {
	if e.Reason == "" {
		return ErrGeminiSafetyBlocked.Error()
	}
	return ErrGeminiSafetyBlocked.Error() + ": " + e.Reason
}

func (e *SafetyBlockError) Unwrap() error { return ErrGeminiSafetyBlocked }

// geminiClient is the production implementation backed by google.golang.org/genai.
type geminiClient struct {
	sdk   *genai.Client
	model string
}

// NewGeminiClient constructs a real Gemini client from configuration. It returns an error
// when the API key or model are missing — callers (e.g., routes wiring) decide whether to
// hard-fail boot or skip the route.
func NewGeminiClient(ctx context.Context, cfg config.GeminiConfig) (GeminiClient, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("ai: gemini api_key is empty")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, errors.New("ai: gemini model is empty")
	}
	backend, err := resolveBackend(cfg.Backend)
	if err != nil {
		return nil, err
	}
	sdk, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: backend,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: failed to construct gemini client: %w", err)
	}
	return &geminiClient{sdk: sdk, model: cfg.Model}, nil
}

// resolveBackend translates the config string into the SDK enum. Empty defaults to the
// Gemini Developer API; "vertex-ai" switches to Vertex express mode (same API-key auth,
// different hostname).
func resolveBackend(name string) (genai.Backend, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "gemini-api", "gemini", "ai-studio":
		return genai.BackendGeminiAPI, nil
	case "vertex-ai", "vertex", "vertexai":
		return genai.BackendVertexAI, nil
	default:
		return 0, fmt.Errorf("ai: unknown gemini backend %q (expected \"gemini-api\" or \"vertex-ai\")", name)
	}
}

func (c *geminiClient) Generate(ctx context.Context, req dto.PolishRequest, temperature float32) ([]string, error) {
	parts := []*genai.Part{genai.NewPartFromText(buildUserPrompt(req))}
	for _, img := range req.Images {
		parts = append(parts, genai.NewPartFromBytes(img.Data, img.MIMEType))
	}

	resp, err := c.sdk.Models.GenerateContent(
		ctx,
		c.model,
		[]*genai.Content{{Parts: parts, Role: genai.RoleUser}},
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
			},
			Temperature:      &temperature,
			ResponseMIMEType: "application/json",
			ResponseSchema:   responseSchema(),
		},
	)
	if err != nil {
		return nil, classifyUpstreamError(err)
	}

	if fb := resp.PromptFeedback; fb != nil && fb.BlockReason != "" {
		return nil, &SafetyBlockError{Reason: string(fb.BlockReason)}
	}
	if len(resp.Candidates) > 0 {
		switch resp.Candidates[0].FinishReason {
		case genai.FinishReasonSafety,
			genai.FinishReasonBlocklist,
			genai.FinishReasonProhibitedContent,
			genai.FinishReasonSPII,
			genai.FinishReasonImageSafety,
			genai.FinishReasonImageProhibitedContent:
			return nil, &SafetyBlockError{Reason: string(resp.Candidates[0].FinishReason)}
		}
	}

	return parseCandidates(resp.Text())
}

// responseSchema constrains the model output to {"candidates":[string,string,string]}.
// The schema is a strong nudge but not a guarantee, so the caller still has to defend
// against malformed or short responses.
func responseSchema() *genai.Schema {
	minItems := int64(3)
	maxItems := int64(3)
	return &genai.Schema{
		Type:     genai.TypeObject,
		Required: []string{"candidates"},
		Properties: map[string]*genai.Schema{
			"candidates": {
				Type:        genai.TypeArray,
				MinItems:    &minItems,
				MaxItems:    &maxItems,
				Description: "Three polished review candidates differing in tone (concise, warm, detailed).",
				Items: &genai.Schema{
					Type:        genai.TypeString,
					Description: "A polished review text in the requested output language.",
				},
			},
		},
	}
}

// parseCandidates decodes the model's JSON output and normalizes the candidate list.
// It tolerates leading/trailing whitespace, fenced code blocks, extra fields, and trims
// each candidate; entries that are empty after trimming are dropped.
func parseCandidates(raw string) ([]string, error) {
	trimmed := stripJSONFence(strings.TrimSpace(raw))
	if trimmed == "" {
		return nil, fmt.Errorf("%w: empty response", ErrGeminiInvalidResponse)
	}
	var payload struct {
		Candidates []string `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGeminiInvalidResponse, err)
	}
	out := make([]string, 0, len(payload.Candidates))
	for _, c := range payload.Candidates {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	return out, nil
}

// stripJSONFence removes a ```json ... ``` markdown fence if the model wraps its output
// in one despite being asked for raw JSON. Returns the input unchanged when no fence
// is present.
func stripJSONFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

// classifyUpstreamError maps an SDK error to one of our sentinels so the handler can
// choose the right HTTP status. The SDK returns genai.APIError by value for HTTP
// failures, carrying the upstream status code.
func classifyUpstreamError(err error) error {
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code == 429 {
			return fmt.Errorf("%w: %v", ErrGeminiRateLimited, err)
		}
	}
	return fmt.Errorf("%w: %v", ErrGeminiUpstream, err)
}
