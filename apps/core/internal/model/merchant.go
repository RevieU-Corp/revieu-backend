package model

import "time"

type Merchant struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Category    string    `gorm:"type:varchar(50)" json:"category"`
	Address     string    `gorm:"type:varchar(255)" json:"address"`
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	CoverImage  string    `gorm:"type:varchar(255)" json:"cover_image"`
	Description string    `gorm:"type:text" json:"description"`
	AvgRating   float32   `gorm:"default:0" json:"avg_rating"`
	ReviewCount int       `gorm:"default:0" json:"review_count"`
	Status      int16     `gorm:"default:0" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m *Merchant) TableName() string {
	return "merchants"
}
