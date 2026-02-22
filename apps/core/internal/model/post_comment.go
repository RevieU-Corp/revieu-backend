package model

import "time"

type PostComment struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID          int64     `gorm:"not null;index" json:"post_id"`
	UserID          int64     `gorm:"not null;index" json:"user_id"`
	ParentCommentID *int64    `gorm:"index" json:"parent_comment_id"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	LikeCount       int       `gorm:"default:0" json:"like_count"`
	Status          int16     `gorm:"default:0" json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	User    *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post    *Post        `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Replies []PostComment `gorm:"foreignKey:ParentCommentID" json:"replies,omitempty"`
}

func (pc *PostComment) TableName() string { return "post_comments" }
