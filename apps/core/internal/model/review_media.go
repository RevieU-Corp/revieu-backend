package model

import "time"

type ReviewMedia struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewID  int64     `gorm:"not null;index" json:"review_id"`
	MediaType string    `gorm:"type:varchar(10);not null" json:"media_type"`
	URL       string    `gorm:"type:varchar(512);not null" json:"url"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (rm *ReviewMedia) TableName() string { return "review_media" }
