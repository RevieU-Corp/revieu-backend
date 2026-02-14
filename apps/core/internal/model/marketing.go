package model

import "time"

type MarketingPost struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID  int64     `gorm:"not null;index" json:"merchant_id"`
	StoreID     *int64    `gorm:"index" json:"store_id"`
	Title       string    `gorm:"type:varchar(100);not null" json:"title"`
	Content     string    `gorm:"type:text" json:"content"`
	Images      string    `gorm:"type:jsonb;default:'[]'" json:"images"`
	PostType    string    `gorm:"type:varchar(20);default:'promotion'" json:"post_type"`
	CouponID    *int64    `gorm:"index" json:"coupon_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	ViewCount   int       `gorm:"default:0" json:"view_count"`
	ClickCount  int       `gorm:"default:0" json:"click_count"`
	Status      string    `gorm:"type:varchar(20);default:'draft'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (mp *MarketingPost) TableName() string { return "marketing_posts" }
