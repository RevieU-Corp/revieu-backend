package dto

import (
	"fmt"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

type Merchant struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	BusinessName string   `json:"businessName"`
	Category     string   `json:"category"`
	Rating       float32  `json:"rating"`
	ReviewCount  int      `json:"reviewCount"`
	Distance     string   `json:"distance"`
	Tags         []string `json:"tags"`
	CoverImage   string   `json:"coverImage"`
}

func FromModel(m model.Merchant) Merchant {
	name := m.Name
	if m.BusinessName != "" {
		name = m.BusinessName
	}
	return Merchant{
		ID:           fmt.Sprintf("%d", m.ID),
		Name:         name,
		BusinessName: m.BusinessName,
		Category:     m.Category,
		Rating:       m.AvgRating,
		ReviewCount:  m.ReviewCount,
		Distance:     "",
		Tags:         []string{},
		CoverImage:   m.CoverImage,
	}
}
