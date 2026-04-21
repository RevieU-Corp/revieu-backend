package service

import (
	"strings"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
)

func ptrFloat(v float64) *float64 { return &v }

func TestBuildUserPrompt_NoContext(t *testing.T) {
	req := dto.PolishRequest{Text: "The coffee was great and the vibe was chill."}
	got := buildUserPrompt(req)

	if strings.Contains(got, "Context:") {
		t.Fatalf("expected no Context block when no context fields set, got:\n%s", got)
	}
	if !strings.Contains(got, "--- DRAFT ---") || !strings.Contains(got, "--- END DRAFT ---") {
		t.Fatalf("expected draft delimiters in output, got:\n%s", got)
	}
	if !strings.Contains(got, req.Text) {
		t.Fatalf("expected draft text verbatim in output, got:\n%s", got)
	}
}

func TestBuildUserPrompt_FullContext(t *testing.T) {
	req := dto.PolishRequest{
		Text:             "Pasta was great, service slow.",
		MerchantName:     "Bella Italia",
		StoreName:        "Downtown",
		BusinessCategory: "restaurant",
		Language:         "en",
		Rating: dto.PolishRating{
			Overall: ptrFloat(4.0),
			Service: ptrFloat(3.0),
			Food:    ptrFloat(5.0),
		},
		Images: []dto.PolishImage{{MIMEType: "image/jpeg", Data: []byte{0xff}}},
	}
	got := buildUserPrompt(req)

	wantSubs := []string{
		"Context:",
		"Merchant name: Bella Italia",
		"Store name: Downtown",
		"Business category: restaurant",
		"Output language: en",
		"User ratings (0-5 scale):",
		"Overall: 4.0",
		"Service: 3.0",
		"Food: 5.0",
		"1 image(s) attached",
		"--- DRAFT ---",
		req.Text,
		"--- END DRAFT ---",
	}
	for _, sub := range wantSubs {
		if !strings.Contains(got, sub) {
			t.Errorf("prompt missing %q\nfull prompt:\n%s", sub, got)
		}
	}

	if strings.Contains(got, "Environment") {
		t.Errorf("unspecified Environment rating should be omitted, got:\n%s", got)
	}
	if strings.Contains(got, "Value") {
		t.Errorf("unspecified Value rating should be omitted, got:\n%s", got)
	}
}

func TestBuildUserPrompt_UnicodeTextPreserved(t *testing.T) {
	text := "咖啡很棒，氛围也很好 ☕️"
	got := buildUserPrompt(dto.PolishRequest{Text: text, Language: "zh"})
	if !strings.Contains(got, text) {
		t.Fatalf("expected unicode draft preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "Output language: zh") {
		t.Fatalf("expected language line, got:\n%s", got)
	}
}

func TestParseCandidates_HappyPath(t *testing.T) {
	raw := `{"candidates":["one","two","three"]}`
	got, err := parseCandidates(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0] != "one" || got[2] != "three" {
		t.Fatalf("unexpected candidates: %v", got)
	}
}

func TestParseCandidates_TrimsAndDropsEmpty(t *testing.T) {
	raw := `{"candidates":["  alpha  ", "", "beta", "   "]}`
	got, err := parseCandidates(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("expected [alpha beta], got %v", got)
	}
}

func TestParseCandidates_FencedOutput(t *testing.T) {
	raw := "```json\n{\"candidates\":[\"a\",\"b\",\"c\"]}\n```"
	got, err := parseCandidates(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 candidates, got %v", got)
	}
}

func TestParseCandidates_MalformedJSON(t *testing.T) {
	_, err := parseCandidates(`{"candidates": [not valid`)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestParseCandidates_Empty(t *testing.T) {
	if _, err := parseCandidates(""); err == nil {
		t.Fatal("expected error for empty response")
	}
	if _, err := parseCandidates("   \n  "); err == nil {
		t.Fatal("expected error for whitespace-only response")
	}
}

func TestParseCandidates_FewerThanThree(t *testing.T) {
	got, err := parseCandidates(`{"candidates":["only one"]}`)
	if err != nil {
		t.Fatalf("parser should not error on short list; service decides: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %v", got)
	}
}

func TestParseCandidates_MoreThanThree(t *testing.T) {
	got, err := parseCandidates(`{"candidates":["a","b","c","d","e"]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("parser should not trim excess; service decides: got %v", got)
	}
}
