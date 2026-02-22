package model

import "time"

type ReviewComment struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewID        int64     `gorm:"not null;index" json:"review_id"`
	UserID          int64     `gorm:"not null;index" json:"user_id"`
	ParentCommentID *int64    `gorm:"index" json:"parent_comment_id"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	IsMerchantReply bool      `gorm:"default:false" json:"is_merchant_reply"`
	UserNickname    string    `gorm:"-" json:"user_nickname,omitempty"`
	UserAvatar      string    `gorm:"-" json:"user_avatar,omitempty"`
	Status          int16     `gorm:"default:0" json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Replies []ReviewComment `gorm:"foreignKey:ParentCommentID" json:"replies,omitempty"`
}

func (r *ReviewComment) TableName() string {
	return "review_comments"
}
