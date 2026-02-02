package service

import (
	"context"
	"testing"

	userdto "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
)

func TestUserServiceProfileAndSettings(t *testing.T) {
	db := testutil.SetupTestDB(t)
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
	if err := svc.UpdateProfile(context.Background(), user.ID, userdto.UpdateProfileRequest{Nickname: &newName}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
}

func TestUserServiceAddressDefault(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := NewUserService(db)
	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	addr, err := svc.CreateAddress(context.Background(), user.ID, userdto.CreateAddressRequest{Name: "A", Phone: "1", Address: "X"})
	if err != nil || !addr.IsDefault {
		t.Fatalf("expected default address")
	}
}
