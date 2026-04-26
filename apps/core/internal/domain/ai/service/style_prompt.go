package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
)

// styleSystemInstruction sets the role and the two non-negotiable rules: write features
// in English (so the same profile can be applied across draft languages) and quote
// snippets verbatim from the inputs (so the polish-side voice anchor stays authentic).
const styleSystemInstruction = "You analyze a user's review-writing voice. " +
	"Given a sequence of reviews they have submitted, output a single JSON object with two top-level keys: " +
	"\"features\" (a structured description of their baseline style) and \"samples\" (1-2 short verbatim excerpts). " +
	"All free-text fields inside features MUST be in English regardless of the input language. " +
	"Each sample's \"snippet\" MUST be quoted verbatim from one of the inputs (no paraphrasing, no translation), 60-200 characters. " +
	"Each sample MUST include the matching review_id from the inputs. " +
	"Schema for features: { tone: string[], formality: integer 1-5, length_preference: \"short\"|\"medium\"|\"long\", " +
	"structure: \"narrative\"|\"listy\"|\"mixed\", emoji_usage: \"none\"|\"rare\"|\"frequent\", " +
	"signature_phrases: string[], frequent_topics: string[], voice_summary: string }. " +
	"Return raw JSON only, no markdown fences."

// styleSnippetMaxLen bounds the snippet length we will accept from the model. The system
// prompt asks for 200 chars, but we trim defensively.
const styleSnippetMaxLen = 200

// styleMaxSamples caps how many samples we keep, regardless of what the model returned.
// Two is enough as a voice anchor; more inflates the polish prompt.
const styleMaxSamples = 2

// buildStyleUserPrompt formats the input reviews as a labeled list the model can quote
// from. Newlines inside content are preserved so the model sees the user's actual line
// structure.
func buildStyleUserPrompt(reviews []ReviewForStyle) string {
	var b strings.Builder
	b.WriteString("Reviews submitted by the user (most recent first):\n\n")
	for i, r := range reviews {
		fmt.Fprintf(&b, "[%d] review_id=%d:\n%s\n---\n", i+1, r.ID, r.Content)
	}
	b.WriteString("\nReturn the JSON object now.")
	return b.String()
}

// parseStyleResponse decodes the model's raw output and returns a sanitized profile.
// It drops samples whose review_id is not in the input set (model fabrication guard) and
// trims each snippet to styleSnippetMaxLen. The features object is taken as-is — the
// polish prompt builder skips empty fields, so loose parsing is safe.
func parseStyleResponse(raw string, inputs []ReviewForStyle) (dto.StyleProfile, []dto.SampleSnippet, error) {
	trimmed := stripJSONFence(strings.TrimSpace(raw))
	if trimmed == "" {
		return dto.StyleProfile{}, nil, fmt.Errorf("%w: empty derive response", ErrGeminiInvalidResponse)
	}

	var payload struct {
		Features dto.StyleProfile    `json:"features"`
		Samples  []dto.SampleSnippet `json:"samples"`
	}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return dto.StyleProfile{}, nil, fmt.Errorf("%w: %v", ErrGeminiInvalidResponse, err)
	}

	allowed := make(map[int64]struct{}, len(inputs))
	for _, r := range inputs {
		allowed[r.ID] = struct{}{}
	}

	cleaned := make([]dto.SampleSnippet, 0, len(payload.Samples))
	for _, s := range payload.Samples {
		snippet := strings.TrimSpace(s.Snippet)
		if snippet == "" {
			continue
		}
		if _, ok := allowed[s.ReviewID]; !ok {
			// Reject snippets the model attached to a fabricated review_id.
			continue
		}
		if len([]rune(snippet)) > styleSnippetMaxLen {
			runes := []rune(snippet)
			snippet = string(runes[:styleSnippetMaxLen])
		}
		cleaned = append(cleaned, dto.SampleSnippet{ReviewID: s.ReviewID, Snippet: snippet})
		if len(cleaned) >= styleMaxSamples {
			break
		}
	}

	return payload.Features, cleaned, nil
}
