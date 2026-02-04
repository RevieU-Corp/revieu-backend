package model

import "time"

type MediaUpload struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UploadURL string    `gorm:"type:varchar(255)" json:"upload_url"`
	FileURL   string    `gorm:"type:varchar(255)" json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *MediaUpload) TableName() string { return "media_uploads" }
