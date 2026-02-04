package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

type Review struct {
	ID         string   `json:"id"`
	MerchantID string   `json:"merchantId"`
	UserID     string   `json:"userId"`
	Rating     float64  `json:"rating"`
	Text       string   `json:"text"`
	Images     []string `json:"images"`
	Tags       []string `json:"tags"`
	CreatedAt  string   `json:"createdAt"`
}

func (r Review) MerchantIDValue() (int64, error) {
	if r.MerchantID == "" {
		return 0, errors.New("merchantId required")
	}
	return strconv.ParseInt(r.MerchantID, 10, 64)
}

func FromModel(m model.Review) Review {
	return Review{
		ID:         strconv.FormatInt(m.ID, 10),
		MerchantID: strconv.FormatInt(m.MerchantID, 10),
		UserID:     strconv.FormatInt(m.UserID, 10),
		Rating:     float64(m.Rating),
		Text:       m.Content,
		Images:     []string{},
		Tags:       []string{},
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
