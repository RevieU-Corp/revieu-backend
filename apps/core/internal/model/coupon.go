package model

import "time"

type Coupon struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID int64     `gorm:"not null;index" json:"merchant_id"`
	Title      string    `gorm:"type:varchar(100);not null" json:"title"`
	Type       string    `gorm:"type:varchar(20);not null" json:"type"`
	Value      string    `gorm:"type:varchar(50)" json:"value"`
	Price      float64   `json:"price"`
	ExpiryDate time.Time `json:"expiry_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Coupon) TableName() string { return "coupons" }
