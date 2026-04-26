package dto

// StyleProfile is the structured English-only voice description that the AI service
// injects into the polish prompt when a user opts to use their saved writing style.
// Every field is optional; renderers must skip blank values rather than read meaning
// into the zero value (e.g. Formality == 0 means "not specified", not "very informal").
type StyleProfile struct {
	Tone             []string `json:"tone,omitempty"`
	Formality        int      `json:"formality,omitempty"`         // 1-5
	LengthPreference string   `json:"length_preference,omitempty"` // short | medium | long
	Structure        string   `json:"structure,omitempty"`         // narrative | listy | mixed
	EmojiUsage       string   `json:"emoji_usage,omitempty"`       // none | rare | frequent
	SignaturePhrases []string `json:"signature_phrases,omitempty"`
	FrequentTopics   []string `json:"frequent_topics,omitempty"`
	VoiceSummary     string   `json:"voice_summary,omitempty"`
}

// SampleSnippet is a verbatim slice of one of the user's past reviews. The polish prompt
// uses these as a "voice anchor" — the model imitates the rhythm/register, but the
// system instruction forbids copying facts from snippets into the new candidates.
type SampleSnippet struct {
	ReviewID int64  `json:"review_id"`
	Snippet  string `json:"snippet"`
}
