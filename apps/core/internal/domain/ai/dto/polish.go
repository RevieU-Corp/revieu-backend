package dto

// PolishRating carries the optional star-rating breakdown supplied with a polish request.
// Each field is a pointer so we can distinguish "not provided" from a zero rating, which
// lets the prompt builder omit unspecified fields from the model context.
type PolishRating struct {
	Overall     *float64
	Service     *float64
	Environment *float64
	Value       *float64
	Food        *float64
}

// PolishImage is one image attachment forwarded to the model as inline binary data.
type PolishImage struct {
	MIMEType string
	Data     []byte
}

// PolishRequest is the validated, multipart-parsed input the AI service receives.
// UserID and UseStyle drive the per-call writing-style personalization; StyleProfile
// and StyleSamples are populated by the AI service after the request reaches it (the
// handler does not look those up). Keeping them on the request keeps the prompt builder
// pure and easy to unit-test.
type PolishRequest struct {
	Text             string
	Images           []PolishImage
	MerchantName     string
	StoreName        string
	BusinessCategory string
	Language         string
	Rating           PolishRating

	// UserID identifies the authenticated caller. Zero means unauthenticated and disables
	// any style lookup regardless of UseStyle.
	UserID int64
	// UseStyle is the per-request toggle the user controls in the polish UI. Defaults to
	// true on the handler side; set to false here if the user opted out for this draft.
	UseStyle bool

	// StyleProfile and StyleSamples are filled in by the AI service before prompt
	// assembly. The prompt builder reads them; nil means no profile, skip injection.
	StyleProfile *StyleProfile
	StyleSamples []SampleSnippet
}

// PolishResponse is the JSON body returned to the client. StyleApplied tells the
// frontend whether the polished candidates actually used the user's saved writing
// style — true only when UseStyle was on, the user has a derived profile, and lookup
// succeeded. Frontends use it to surface a "applied your writing style" hint.
type PolishResponse struct {
	Candidates   []string `json:"candidates"`
	StyleApplied bool     `json:"style_applied"`
}
