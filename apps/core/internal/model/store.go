package model

import "time"

type Store struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID    int64     `gorm:"not null;index" json:"merchant_id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	Address       string    `gorm:"type:varchar(255)" json:"address"`
	City          string    `gorm:"type:varchar(100)" json:"city"`
	State         string    `gorm:"type:varchar(100)" json:"state"`
	ZipCode       string    `gorm:"type:varchar(20)" json:"zip_code"`
	Country       string    `gorm:"type:varchar(50)" json:"country"`
	Phone         string    `gorm:"type:varchar(50)" json:"phone"`
	Website       string    `gorm:"type:varchar(255)" json:"website"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	CoverImageURL string    `gorm:"type:varchar(255)" json:"cover_image_url"`
	Images        string    `gorm:"type:jsonb;default:'[]'" json:"images"`
	AvgRating     float32   `gorm:"default:0" json:"avg_rating"`
	ReviewCount   int       `gorm:"default:0" json:"review_count"`
	FollowerCount int       `gorm:"default:0" json:"follower_count"`
	Status        int16     `gorm:"default:0" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Merchant   *Merchant   `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Categories []Category  `gorm:"many2many:store_categories" json:"categories,omitempty"`
	Hours      []StoreHour `gorm:"foreignKey:StoreID" json:"hours,omitempty"`
}

func (s *Store) TableName() string { return "stores" }

type StoreHour struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID   int64  `gorm:"not null;index" json:"store_id"`
	DayOfWeek int16  `gorm:"not null" json:"day_of_week"`
	OpenTime  string `gorm:"type:varchar(10)" json:"open_time"`
	CloseTime string `gorm:"type:varchar(10)" json:"close_time"`
	IsClosed  bool   `gorm:"default:false" json:"is_closed"`
}

func (sh *StoreHour) TableName() string { return "store_hours" }
