package service

import (
	"fmt"
	"strings"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
)

// systemInstruction is the role / output-contract preamble sent as Gemini's SystemInstruction.
// It is intentionally terse: the structured-output schema enforces the JSON shape, so the prompt
// only needs to anchor the role, the don't-invent-facts rule, and (when present) the
// writing-style baseline-vs-draft conflict rule.
const systemInstruction = "You are an editor that polishes user-written venue reviews. " +
	"Preserve the user's facts, sentiment, and rating; correct grammar and improve flow; " +
	"never invent details that are not present in the draft text or the attached images. " +
	"Match the language requested by the user, or the language of the draft if none is requested. " +
	"Return JSON that conforms to the response schema, producing exactly three distinct candidates " +
	"that differ in tone: one concise, one warm, one detailed. " +
	"If the user message includes a 'User writing style' block, treat it as the user's baseline voice " +
	"and let it inform word choice, rhythm, and default tone — but never override the sentiment, " +
	"formality, or specific content choices the user clearly made in this draft. " +
	"Reference snippets, when provided, are voice anchors only: imitate their register, never copy their facts."

// buildUserPrompt assembles the user-turn text: optional context block, optional
// writing-style block, then the draft. Image parts are added separately by the Gemini
// client wrapper. The function is pure so it can be golden-tested without touching the SDK.
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

	if block := styleBlock(req.StyleProfile, req.StyleSamples); block != "" {
		b.WriteString(block)
		b.WriteByte('\n')
	}

	b.WriteString("--- DRAFT ---\n")
	b.WriteString(req.Text)
	b.WriteString("\n--- END DRAFT ---")

	return b.String()
}

// styleBlock renders the writing-style section when a profile is provided. It returns an
// empty string for nil/empty profiles so the caller can append unconditionally. Each
// feature line is only emitted when present, so a half-derived profile still produces a
// usable block instead of "Tone: , Formality 0/5" noise.
func styleBlock(profile *dto.StyleProfile, samples []dto.SampleSnippet) string {
	if profile == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("User writing style (in English; apply to whatever language the draft uses):\n")
	wrote := false
	if profile.VoiceSummary != "" {
		fmt.Fprintf(&b, "- Voice: %s\n", profile.VoiceSummary)
		wrote = true
	}
	if toneLine := formatToneLine(profile); toneLine != "" {
		fmt.Fprintf(&b, "- %s\n", toneLine)
		wrote = true
	}
	if profile.Structure != "" || profile.EmojiUsage != "" {
		var bits []string
		if profile.Structure != "" {
			bits = append(bits, "Structure: "+profile.Structure)
		}
		if profile.EmojiUsage != "" {
			bits = append(bits, "Emoji: "+profile.EmojiUsage)
		}
		fmt.Fprintf(&b, "- %s\n", strings.Join(bits, ". "))
		wrote = true
	}
	if len(profile.SignaturePhrases) > 0 {
		fmt.Fprintf(&b, "- Common phrases: %s\n", strings.Join(profile.SignaturePhrases, ", "))
		wrote = true
	}
	if len(profile.FrequentTopics) > 0 {
		fmt.Fprintf(&b, "- Often comments on: %s\n", strings.Join(profile.FrequentTopics, ", "))
		wrote = true
	}

	if len(samples) > 0 {
		b.WriteString("Style reference snippets (verbatim from past reviews; for voice only, do not copy their facts):\n")
		for i, s := range samples {
			fmt.Fprintf(&b, "%d. %q\n", i+1, s.Snippet)
		}
		wrote = true
	}

	if !wrote {
		return ""
	}
	return b.String()
}

// formatToneLine combines tone, formality, and length into a single bullet so the
// block stays compact. Returns "" when none of the three are populated.
func formatToneLine(profile *dto.StyleProfile) string {
	var bits []string
	if len(profile.Tone) > 0 {
		bits = append(bits, "Tone: "+strings.Join(profile.Tone, ", "))
	}
	if profile.Formality > 0 {
		bits = append(bits, fmt.Sprintf("Formality %d/5", profile.Formality))
	}
	if profile.LengthPreference != "" {
		bits = append(bits, "Length: "+profile.LengthPreference)
	}
	return strings.Join(bits, ". ")
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
