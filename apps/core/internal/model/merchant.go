package model

import "time"

type Merchant struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID             *int64     `gorm:"index" json:"user_id"`
	Name               string     `gorm:"type:varchar(100);not null" json:"name"`
	BusinessName       string     `gorm:"type:varchar(100)" json:"business_name"`
	BusinessType       string     `gorm:"type:varchar(50)" json:"business_type"`
	Category           string     `gorm:"type:varchar(50)" json:"category"`
	LogoURL            string     `gorm:"type:varchar(255)" json:"logo_url"`
	Address            string     `gorm:"type:varchar(255)" json:"address"`
	Phone              string     `gorm:"type:varchar(20)" json:"phone"`
	ContactPhone       string     `gorm:"type:varchar(20)" json:"contact_phone"`
	ContactEmail       string     `gorm:"type:varchar(255)" json:"contact_email"`
	WebsiteURL         string     `gorm:"type:varchar(255)" json:"website_url"`
	SocialLinks        string     `gorm:"type:jsonb;default:'{}'" json:"social_links"`
	CoverImage         string     `gorm:"type:varchar(255)" json:"cover_image"`
	Description        string     `gorm:"type:text" json:"description"`
	AvgRating          float32    `gorm:"default:0" json:"avg_rating"`
	ReviewCount        int        `gorm:"default:0" json:"review_count"`
	FollowerCount      int        `gorm:"default:0" json:"follower_count"`
	TotalStores        int        `gorm:"default:0" json:"total_stores"`
	TotalReviews       int        `gorm:"default:0" json:"total_reviews"`
	VerificationStatus string     `gorm:"type:varchar(20);default:'unverified'" json:"verification_status"`
	VerifiedAt         *time.Time `json:"verified_at"`
	Status             int16      `gorm:"default:0" json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Stores []Store `gorm:"foreignKey:MerchantID" json:"stores,omitempty"`
}

func (m *Merchant) TableName() string {
	return "merchants"
}
