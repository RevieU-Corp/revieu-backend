package model

import "time"

type UserPrivacy struct {
	UserID   int64 `gorm:"primaryKey" json:"user_id"`
	IsPublic bool  `gorm:"default:true" json:"is_public"`
}

func (u *UserPrivacy) TableName() string {
	return "user_privacies"
}

type UserNotification struct {
	UserID       int64 `gorm:"primaryKey" json:"user_id"`
	PushEnabled  bool  `gorm:"default:true" json:"push_enabled"`
	EmailEnabled bool  `gorm:"default:true" json:"email_enabled"`
}

func (u *UserNotification) TableName() string {
	return "user_notifications"
}

type AccountDeletion struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"not null;uniqueIndex" json:"user_id"`
	Reason      string    `gorm:"type:varchar(255)" json:"reason"`
	ScheduledAt time.Time `gorm:"not null" json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (a *AccountDeletion) TableName() string {
	return "account_deletions"
}
