package model

import "time"

type Voucher struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code           string     `gorm:"type:varchar(50);uniqueIndex" json:"code"`
	CouponID       int64      `gorm:"not null;index" json:"coupon_id"`
	UserID         int64      `gorm:"not null;index" json:"user_id"`
	PackageID      *int64     `gorm:"index" json:"package_id"`
	OrderID        *int64     `gorm:"index" json:"order_id"`
	MerchantID     *int64     `gorm:"index" json:"merchant_id"`
	Status         string     `gorm:"type:varchar(20);not null" json:"status"`
	ExpiryDate     time.Time  `json:"expiry_date"`
	ValidFrom      *time.Time `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
	RedeemedAt     *time.Time `json:"redeemed_at"`
	RedeemedBy     *int64     `json:"redeemed_by"`
	RedemptionNote string     `gorm:"type:text" json:"redemption_note"`
	CouponTitle    string     `gorm:"-" json:"coupon_title,omitempty"`
	MerchantName   string     `gorm:"-" json:"merchant_name,omitempty"`
	QRCode         string     `gorm:"type:varchar(255)" json:"qr_code"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (v *Voucher) TableName() string { return "vouchers" }
