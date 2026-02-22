package model

import "time"

type BrowsingHistory struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	StoreID   *int64    `gorm:"index" json:"store_id"`
	ReviewID  *int64    `gorm:"index" json:"review_id"`
	PostID    *int64    `gorm:"index" json:"post_id"`
	ViewedAt  time.Time `json:"viewed_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (bh *BrowsingHistory) TableName() string { return "browsing_histories" }
