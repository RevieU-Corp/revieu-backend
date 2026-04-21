package service

import (
	"fmt"
	"strings"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
)

// systemInstruction is the role / output-contract preamble sent as Gemini's SystemInstruction.
// It is intentionally terse: the structured-output schema enforces the JSON shape, so the prompt
// only needs to anchor the role and the don't-invent-facts rule.
const systemInstruction = "You are an editor that polishes user-written venue reviews. " +
	"Preserve the user's facts, sentiment, and rating; correct grammar and improve flow; " +
	"never invent details that are not present in the draft text or the attached images. " +
	"Match the language requested by the user, or the language of the draft if none is requested. " +
	"Return JSON that conforms to the response schema, producing exactly three distinct candidates " +
	"that differ in tone: one concise, one warm, one detailed."

// buildUserPrompt assembles the user-turn text: optional context block, then the draft.
// Image parts are added separately by the Gemini client wrapper. The function is pure so
// it can be golden-tested without touching the SDK.
func buildUserPrompt(req dto.PolishRequest) string {
	var b strings.Builder

	contextLines := contextBlockLines(req)
	if len(contextLines) > 0 {
		b.WriteString("Context:\n")
		for _, line := range contextLines {
			b.WriteString("- ")
			b.WriteString(line)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	b.WriteString("--- DRAFT ---\n")
	b.WriteString(req.Text)
	b.WriteString("\n--- END DRAFT ---")

	return b.String()
}

func contextBlockLines(req dto.PolishRequest) []string {
	var lines []string
	if req.MerchantName != "" {
		lines = append(lines, "Merchant name: "+req.MerchantName)
	}
	if req.StoreName != "" {
		lines = append(lines, "Store name: "+req.StoreName)
	}
	if req.BusinessCategory != "" {
		lines = append(lines, "Business category: "+req.BusinessCategory)
	}
	if req.Language != "" {
		lines = append(lines, "Output language: "+req.Language)
	}

	ratingLines := ratingContextLines(req.Rating)
	if len(ratingLines) > 0 {
		lines = append(lines, "User ratings (0-5 scale):")
		lines = append(lines, ratingLines...)
	}
	if len(req.Images) > 0 {
		lines = append(lines, fmt.Sprintf("%d image(s) attached for visual reference", len(req.Images)))
	}
	return lines
}

func ratingContextLines(r dto.PolishRating) []string {
	var lines []string
	add := func(label string, v *float64) {
		if v != nil {
			lines = append(lines, fmt.Sprintf("  %s: %.1f", label, *v))
		}
	}
	add("Overall", r.Overall)
	add("Service", r.Service)
	add("Environment", r.Environment)
	add("Value", r.Value)
	add("Food", r.Food)
	return lines
}
