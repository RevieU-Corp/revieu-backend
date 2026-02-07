package dto

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

type Review struct {
	ID         string   `json:"id"`
	MerchantID string   `json:"merchantId"`
	VenueID    string   `json:"venueId"`
	UserID     string   `json:"userId"`
	Rating     float64  `json:"rating"`
	Text       string   `json:"text"`
	Images     []string `json:"images"`
	Tags       []string `json:"tags"`
	VisitDate  string   `json:"visitDate"`
	CreatedAt  string   `json:"createdAt"`
}

// CommentRequest is the request body for adding a review comment.
type CommentRequest struct {
	Text string `json:"text" binding:"required"`
}

func (r Review) MerchantIDValue() (int64, error) {
	if r.MerchantID == "" {
		return 0, errors.New("merchantId required")
	}
	return strconv.ParseInt(r.MerchantID, 10, 64)
}

func (r Review) VenueIDValue() (int64, error) {
	if r.VenueID == "" {
		return 0, errors.New("venueId required")
	}
	return strconv.ParseInt(r.VenueID, 10, 64)
}

func (r Review) VisitDateValue() (time.Time, error) {
	if r.VisitDate == "" {
		return time.Now(), nil
	}
	return time.Parse("2006-01-02", r.VisitDate)
}

func FromModel(m model.Review) Review {
	var images []string
	if m.Images != "" {
		_ = json.Unmarshal([]byte(m.Images), &images)
	}
	if images == nil {
		images = []string{}
	}
	return Review{
		ID:         strconv.FormatInt(m.ID, 10),
		MerchantID: strconv.FormatInt(m.MerchantID, 10),
		VenueID:    strconv.FormatInt(m.VenueID, 10),
		UserID:     strconv.FormatInt(m.UserID, 10),
		Rating:     float64(m.Rating),
		Text:       m.Content,
		Images:     images,
		Tags:       []string{},
		VisitDate:  m.VisitDate.Format("2006-01-02"),
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
	}
}

func FromModels(items []model.Review) []Review {
	out := make([]Review, 0, len(items))
	for _, item := range items {
		out = append(out, FromModel(item))
	}
	return out
}
