package model

import "time"

type Package struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID         int64     `gorm:"not null;index" json:"merchant_id"`
	StoreID            *int64    `gorm:"index" json:"store_id"`
	Title              string    `gorm:"type:varchar(100);not null" json:"title"`
	Description        string    `gorm:"type:text" json:"description"`
	ImageURL           string    `gorm:"type:varchar(255)" json:"image_url"`
	OriginalPrice      float64   `gorm:"type:numeric(10,2)" json:"original_price"`
	SalePrice          float64   `gorm:"type:numeric(10,2)" json:"sale_price"`
	DiscountPercentage float64   `gorm:"type:numeric(5,2)" json:"discount_percentage"`
	TotalQuantity      int       `gorm:"default:0" json:"total_quantity"`
	SoldCount          int       `gorm:"default:0" json:"sold_count"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
	Terms              string    `gorm:"type:text" json:"terms"`
	Status             string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Coupons  []Coupon  `gorm:"foreignKey:PackageID" json:"coupons,omitempty"`
}

func (p *Package) TableName() string { return "packages" }
