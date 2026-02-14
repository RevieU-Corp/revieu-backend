package model

import "time"

type Review struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64     `gorm:"not null;index" json:"user_id"`
	VenueID       int64     `gorm:"not null;index" json:"venue_id"`
	MerchantID    int64     `gorm:"not null;index" json:"merchant_id"`
	Rating        float32   `gorm:"not null" json:"rating"`
	RatingEnv     *float32  `json:"rating_env"`
	RatingService *float32  `json:"rating_service"`
	RatingValue   *float32  `json:"rating_value"`
	Content       string    `gorm:"type:text" json:"content"`
	Images        string    `gorm:"type:jsonb;default:'[]'" json:"images"`
	VisitDate     time.Time `gorm:"not null;type:date" json:"visit_date"`
	AvgCost       *int      `json:"avg_cost"`
	LikeCount     int       `gorm:"default:0" json:"like_count"`
	Status        int16     `gorm:"default:0" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Tags     []Tag     `gorm:"many2many:review_tags" json:"tags,omitempty"`
}

func (r *Review) TableName() string {
	return "reviews"
}
