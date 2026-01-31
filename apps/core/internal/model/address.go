package model

import "time"

type UserAddress struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;index" json:"user_id"`
	Name       string    `gorm:"type:varchar(50);not null" json:"name"`
	Phone      string    `gorm:"type:varchar(20);not null" json:"phone"`
	Province   string    `gorm:"type:varchar(50)" json:"province"`
	City       string    `gorm:"type:varchar(50)" json:"city"`
	District   string    `gorm:"type:varchar(50)" json:"district"`
	Address    string    `gorm:"type:varchar(255);not null" json:"address"`
	PostalCode string    `gorm:"type:varchar(20)" json:"postal_code"`
	IsDefault  bool      `gorm:"default:false" json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (a *UserAddress) TableName() string {
	return "user_addresses"
}
