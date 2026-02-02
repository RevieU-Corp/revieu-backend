package model

import "time"

type Post struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;index" json:"user_id"`
	MerchantID *int64    `gorm:"index" json:"merchant_id"`
	Title      string    `gorm:"type:varchar(100)" json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Images     string    `gorm:"type:jsonb;default:'[]'" json:"images"`
	LikeCount  int       `gorm:"default:0" json:"like_count"`
	ViewCount  int       `gorm:"default:0" json:"view_count"`
	Status     int16     `gorm:"default:0" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Tags     []Tag     `gorm:"many2many:post_tags" json:"tags,omitempty"`
}

func (p *Post) TableName() string {
	return "posts"
}
