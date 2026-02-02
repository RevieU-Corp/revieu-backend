package testutil

import (
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// SetupTestDB creates an in-memory sqlite DB with schema migrations.
func SetupTestDB(t *testing.T) *gorm.DB {
    t.Helper()

    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: nil})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    if err := db.AutoMigrate(
        &model.User{},
        &model.UserAuth{},
        &model.UserProfile{},
        &model.EmailVerification{},
        &model.Merchant{},
        &model.Tag{},
        &model.Post{},
        &model.Review{},
        &model.UserFollow{},
        &model.MerchantFollow{},
        &model.Like{},
        &model.Favorite{},
        &model.UserAddress{},
        &model.UserPrivacy{},
        &model.UserNotification{},
        &model.AccountDeletion{},
    ); err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }

    return db
}
