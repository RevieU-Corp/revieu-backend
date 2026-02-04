package model

import "time"

type Payment struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Amount     float64   `json:"amount"`
	Currency   string    `gorm:"type:varchar(10)" json:"currency"`
	Status     string    `gorm:"type:varchar(20)" json:"status"`
	CouponID   *int64    `json:"coupon_id"`
	MerchantID *int64    `json:"merchant_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (p *Payment) TableName() string { return "payments" }
