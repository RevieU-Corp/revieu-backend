package model

import "time"

type MerchantAnalytics struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID      int64     `gorm:"not null;index" json:"merchant_id"`
	StoreID         *int64    `gorm:"index" json:"store_id"`
	Date            time.Time `gorm:"type:date;not null" json:"date"`
	TotalViews      int       `gorm:"default:0" json:"total_views"`
	UniqueVisitors  int       `gorm:"default:0" json:"unique_visitors"`
	ReviewsReceived int       `gorm:"default:0" json:"reviews_received"`
	CouponsRedeemed int       `gorm:"default:0" json:"coupons_redeemed"`
	Revenue         float64   `gorm:"type:numeric(12,2);default:0" json:"revenue"`
	AvgRating       float32   `gorm:"default:0" json:"avg_rating"`
	CreatedAt       time.Time `json:"created_at"`
}

func (ma *MerchantAnalytics) TableName() string { return "merchant_analytics" }
