package model

import "time"

type Tag struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Type        string    `gorm:"type:varchar(20);default:'general'" json:"type"`
	PostCount   int       `gorm:"default:0" json:"post_count"`
	ReviewCount int       `gorm:"default:0" json:"review_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func (t *Tag) TableName() string {
	return "tags"
}
