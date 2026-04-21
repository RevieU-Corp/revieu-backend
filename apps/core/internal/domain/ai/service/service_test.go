package service

import (
	"context"
	"errors"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
)

// fakeGeminiClient is a test double that returns the scripted responses in order.
// Each Generate call pops the next response; if the script is exhausted it returns a marker error.
type fakeGeminiClient struct {
	scripted []fakeResponse
	calls    []fakeCall
}

type fakeResponse struct {
	candidates []string
	err        error
}

type fakeCall struct {
	req         dto.PolishRequest
	temperature float32
}

func (f *fakeGeminiClient) Generate(_ context.Context, req dto.PolishRequest, temperature float32) ([]string, error) {
	f.calls = append(f.calls, fakeCall{req: req, temperature: temperature})
	if len(f.scripted) == 0 {
		return nil, errors.New("fake: script exhausted")
	}
	resp := f.scripted[0]
	f.scripted = f.scripted[1:]
	return resp.candidates, resp.err
}

func newService(fake *fakeGeminiClient) *AIService {
	return NewAIService(fake, config.GeminiConfig{Model: "test", TimeoutSeconds: 1})
}

func validRequest() dto.PolishRequest {
	return dto.PolishRequest{Text: "nice cafe, slow service"}
}

func TestPolishReview_HappyPath(t *testing.T) {
	fake := &fakeGeminiClient{scripted: []fakeResponse{
		{candidates: []string{"one", "two", "three"}},
	}}
	s := newService(fake)

	resp, err := s.PolishReview(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(resp.Candidates))
	}
	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.calls))
	}
	if fake.calls[0].temperature != primaryTemperature {
		t.Fatalf("expected primary temperature, got %f", fake.calls[0].temperature)
	}
}

func TestPolishReview_RetriesOnShortResponse(t *testing.T) {
	fake := &fakeGeminiClient{scripted: []fakeResponse{
		{candidates: []string{"only one"}},
		{candidates: []string{"alpha", "beta", "gamma"}},
	}}
	s := newService(fake)

	resp, err := s.PolishReview(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Candidates) != 3 || resp.Candidates[0] != "alpha" {
		t.Fatalf("expected retry result, got %v", resp.Candidates)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("expected 2 calls (primary + retry), got %d", len(fake.calls))
	}
	if fake.calls[1].temperature != retryTemperature {
		t.Fatalf("expected retry temperature, got %f", fake.calls[1].temperature)
	}
}

func TestPolishReview_TrimsExcessCandidates(t *testing.T) {
	fake := &fakeGeminiClient{scripted: []fakeResponse{
		{candidates: []string{"a", "b", "c", "d", "e"}},
	}}
	s := newService(fake)

	resp, err := s.PolishReview(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Candidates) != 3 {
		t.Fatalf("expected exactly 3 candidates, got %d", len(resp.Candidates))
	}
	if resp.Candidates[0] != "a" || resp.Candidates[2] != "c" {
		t.Fatalf("expected first three, got %v", resp.Candidates)
	}
}

func TestPolishReview_InsufficientAfterRetry(t *testing.T) {
	fake := &fakeGeminiClient{scripted: []fakeResponse{
		{candidates: []string{"one"}},
		{candidates: []string{"two", "three"}},
	}}
	s := newService(fake)

	_, err := s.PolishReview(context.Background(), validRequest())
	if !errors.Is(err, ErrInsufficientCandidates) {
		t.Fatalf("expected ErrInsufficientCandidates, got %v", err)
	}
}

func TestPolishReview_PassesThroughUpstreamError(t *testing.T) {
	boom := errors.New("boom")
	wrapped := errors.Join(ErrGeminiUpstream, boom)
	fake := &fakeGeminiClient{scripted: []fakeResponse{{err: wrapped}}}
	s := newService(fake)

	_, err := s.PolishReview(context.Background(), validRequest())
	if !errors.Is(err, ErrGeminiUpstream) {
		t.Fatalf("expected upstream error passthrough, got %v", err)
	}
}

func TestPolishReview_SafetyBlockSurfacesOnFirstCall(t *testing.T) {
	blocked := &SafetyBlockError{Reason: "HARM_CATEGORY_DANGEROUS_CONTENT"}
	fake := &fakeGeminiClient{scripted: []fakeResponse{{err: blocked}}}
	s := newService(fake)

	_, err := s.PolishReview(context.Background(), validRequest())
	var sb *SafetyBlockError
	if !errors.As(err, &sb) {
		t.Fatalf("expected SafetyBlockError, got %v", err)
	}
	if sb.Reason != "HARM_CATEGORY_DANGEROUS_CONTENT" {
		t.Fatalf("unexpected reason: %s", sb.Reason)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("expected no retry after safety block, got %d calls", len(fake.calls))
	}
}

func TestPolishReview_RejectsEmptyText(t *testing.T) {
	s := newService(&fakeGeminiClient{})
	_, err := s.PolishReview(context.Background(), dto.PolishRequest{Text: ""})
	if !errors.Is(err, ErrEmptyText) {
		t.Fatalf("expected ErrEmptyText, got %v", err)
	}
}

func TestPolishReview_NilClientReturnsError(t *testing.T) {
	s := NewAIService(nil, config.GeminiConfig{Model: "test"})
	_, err := s.PolishReview(context.Background(), validRequest())
	if !errors.Is(err, ErrGeminiUpstream) {
		t.Fatalf("expected ErrGeminiUpstream when client is nil, got %v", err)
	}
}
