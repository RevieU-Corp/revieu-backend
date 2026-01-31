package model

import "time"

type UserFollow struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FollowerID  int64     `gorm:"not null;uniqueIndex:idx_user_follow" json:"follower_id"`
	FollowingID int64     `gorm:"not null;uniqueIndex:idx_user_follow" json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func (u *UserFollow) TableName() string {
	return "user_follows"
}

type MerchantFollow struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;uniqueIndex:idx_merchant_follow" json:"user_id"`
	MerchantID int64     `gorm:"not null;uniqueIndex:idx_merchant_follow" json:"merchant_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (m *MerchantFollow) TableName() string {
	return "merchant_follows"
}
