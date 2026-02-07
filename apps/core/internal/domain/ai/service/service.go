package service

type SuggestionsRequest struct {
	OverallRating    float64 `json:"overallRating"`
	BusinessCategory string  `json:"businessCategory"`
	CurrentText      string  `json:"currentText"`
	MerchantName     string  `json:"merchantName"`
}

type SuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
}

type AIService struct{}

func NewAIService() *AIService { return &AIService{} }

func (s *AIService) Suggestions(req SuggestionsRequest) SuggestionsResponse {
	name := req.MerchantName
	if name == "" {
		name = "this place"
	}
	return SuggestionsResponse{Suggestions: []string{
		"Highlight what you liked about " + name + ".",
		"Mention any standout service or atmosphere details.",
		"Add one concrete example to support your rating.",
	}}
}
