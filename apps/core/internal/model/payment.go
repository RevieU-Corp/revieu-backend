package model

import "time"

type Payment struct {
	ID               int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Amount           float64    `json:"amount"`
	Currency         string     `gorm:"type:varchar(10)" json:"currency"`
	Status           string     `gorm:"type:varchar(20)" json:"status"`
	CouponID         *int64     `json:"coupon_id"`
	MerchantID       *int64     `json:"merchant_id"`
	OrderID          *int64     `gorm:"index" json:"order_id"`
	UserID           *int64     `gorm:"index" json:"user_id"`
	PaymentMethod    string     `gorm:"type:varchar(30)" json:"payment_method"`
	PaymentSessionID string     `gorm:"type:varchar(255)" json:"payment_session_id"`
	PaidAt           *time.Time `json:"paid_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (p *Payment) TableName() string { return "payments" }
