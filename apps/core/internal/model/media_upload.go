package model

import "time"

type MediaUpload struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      string    `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UserID    int64     `gorm:"index;not null" json:"user_id"`
	ObjectKey string    `gorm:"type:varchar(512);not null" json:"object_key"`
	FileURL   string    `gorm:"type:varchar(512)" json:"file_url"`
	Status    string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *MediaUpload) TableName() string { return "media_uploads" }
