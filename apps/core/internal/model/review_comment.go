package model

import "time"

type ReviewComment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewID  int64     `gorm:"not null;index" json:"review_id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *ReviewComment) TableName() string {
	return "review_comments"
}
