package model

import "time"

type Voucher struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Code       string    `gorm:"type:varchar(50);uniqueIndex" json:"code"`
	CouponID   int64     `gorm:"not null;index" json:"coupon_id"`
	UserID     int64     `gorm:"not null;index" json:"user_id"`
	Status     string    `gorm:"type:varchar(20);not null" json:"status"`
	ExpiryDate time.Time `json:"expiry_date"`
	QRCode     string    `gorm:"type:varchar(255)" json:"qr_code"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (v *Voucher) TableName() string { return "vouchers" }
