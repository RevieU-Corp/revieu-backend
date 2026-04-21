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
type PolishRequest struct {
	Text             string
	Images           []PolishImage
	MerchantName     string
	StoreName        string
	BusinessCategory string
	Language         string
	Rating           PolishRating
}

// PolishResponse is the JSON body returned to the client: exactly three polished
// candidate review texts.
type PolishResponse struct {
	Candidates []string `json:"candidates"`
}
