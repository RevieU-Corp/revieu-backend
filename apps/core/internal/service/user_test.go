package service

import (
	"context"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

func TestUserServiceProfileAndSettings(t *testing.T) {
	db := setupTestDB(t)
	svc := NewUserService(db)
	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.UserProfile{UserID: user.ID, Nickname: "n"}).Error; err != nil {
		t.Fatal(err)
	}

	prof, err := svc.GetProfile(context.Background(), user.ID)
	if err != nil || prof.UserID != user.ID {
		t.Fatalf("profile failed: %v", err)
	}

	newName := "new"
	if err := svc.UpdateProfile(context.Background(), user.ID, dto.UpdateProfileRequest{Nickname: &newName}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
}

func TestUserServiceAddressDefault(t *testing.T) {
	db := setupTestDB(t)
	svc := NewUserService(db)
	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	addr, err := svc.CreateAddress(context.Background(), user.ID, dto.CreateAddressRequest{Name: "A", Phone: "1", Address: "X"})
	if err != nil || !addr.IsDefault {
		t.Fatalf("expected default address")
	}
}
