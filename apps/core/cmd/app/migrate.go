package main

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"gorm.io/gorm"
)

func migrationModels() []interface{} {
	return []interface{}{
		// User system
		&model.User{},
		&model.UserAuth{},
		&model.UserProfile{},
		&model.EmailVerification{},
		&model.RefreshToken{},
		// Social
		&model.UserFollow{},
		&model.MerchantFollow{},
		&model.Like{},
		&model.Favorite{},
		// User settings
		&model.UserAddress{},
		&model.UserPrivacy{},
		&model.UserNotification{},
		&model.AccountDeletion{},
		// Merchant & Store
		&model.Merchant{},
		&model.Category{},
		&model.StoreCategory{},
		&model.Store{},
		&model.StoreHour{},
		// Tags
		&model.Tag{},
		// Content
		&model.Review{},
		&model.ReviewMedia{},
		&model.ReviewComment{},
		&model.Post{},
		&model.PostComment{},
		// Commerce
		&model.Package{},
		&model.Coupon{},
		&model.Order{},
		&model.Voucher{},
		&model.Payment{},
		// Media
		&model.MediaUpload{},
		// Messaging
		&model.Conversation{},
		&model.ConversationParticipant{},
		&model.Message{},
		// Merchant verification
		&model.MerchantVerification{},
		// Marketing & Analytics
		&model.MarketingPost{},
		&model.MerchantAnalytics{},
		// Notifications
		&model.Notification{},
		// Reports & Admin
		&model.Report{},
		&model.AdminAuditLog{},
		// Browsing History
		&model.BrowsingHistory{},
	}
}

func runAutoMigrate(db *gorm.DB, enabled bool) error {
	if !enabled {
		return nil
	}
	return db.AutoMigrate(migrationModels()...)
}
