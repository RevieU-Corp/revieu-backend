package model

import "time"

type Like struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;uniqueIndex:idx_like" json:"user_id"`
	TargetType string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_like" json:"target_type"`
	TargetID   int64     `gorm:"not null;uniqueIndex:idx_like" json:"target_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (l *Like) TableName() string {
	return "likes"
}

type Favorite struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;uniqueIndex:idx_favorite" json:"user_id"`
	TargetType string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_favorite" json:"target_type"`
	TargetID   int64     `gorm:"not null;uniqueIndex:idx_favorite" json:"target_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (f *Favorite) TableName() string {
	return "favorites"
}
