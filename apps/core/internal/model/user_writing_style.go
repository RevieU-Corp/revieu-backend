package model

import "time"

// UserWritingStyle stores a per-user writing-voice profile derived from their submitted
// reviews. It is consulted by the AI polish flow to personalize candidates and refreshed
// asynchronously after every N submitted reviews. LastDerivedAt is nil until the first
// successful derivation — callers must treat that case as "no profile yet" and skip
// injection rather than reading the empty Features blob.
type UserWritingStyle struct {
	UserID                 int64      `gorm:"primaryKey" json:"user_id"`
	Features               string     `gorm:"type:jsonb;not null;default:'{}'" json:"features"`
	Samples                string     `gorm:"type:jsonb;not null;default:'[]'" json:"samples"`
	ReviewsSinceLastDerive int        `gorm:"not null;default:0" json:"reviews_since_last_derive"`
	DerivedFromReviewCount int        `gorm:"not null;default:0" json:"derived_from_review_count"`
	LastDerivedAt          *time.Time `json:"last_derived_at"`
	LastDeriveStartedAt    *time.Time `json:"last_derive_started_at"`
	UpdatedAt              time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s *UserWritingStyle) TableName() string {
	return "user_writing_styles"
}
