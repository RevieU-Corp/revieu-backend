package model

import "time"

type Order struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;index" json:"user_id"`
	CouponID   *int64    `gorm:"index" json:"coupon_id"`
	PackageID  *int64    `gorm:"index" json:"package_id"`
	MerchantID *int64    `gorm:"index" json:"merchant_id"`
	StoreID    *int64    `gorm:"index" json:"store_id"`
	Quantity   int       `gorm:"default:1" json:"quantity"`
	TotalPrice float64   `gorm:"type:numeric(10,2)" json:"total_price"`
	Status     string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Note       string    `gorm:"type:text" json:"note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Coupon *Coupon `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
}

func (o *Order) TableName() string { return "orders" }
