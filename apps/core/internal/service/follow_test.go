package service

import (
	"context"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

func TestFollowServiceUserFollow(t *testing.T) {
	db := setupTestDB(t)
	svc := NewFollowService(db)
	u1 := model.User{Role: "user", Status: 0}
	u2 := model.User{Role: "user", Status: 0}
	db.Create(&u1)
	db.Create(&u2)
	if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
		t.Fatal(err)
	}
}

func TestFollowUpdatesCounts(t *testing.T) {
	db := setupTestDB(t)
	svc := NewFollowService(db)
	u1 := model.User{Role: "user", Status: 0}
	u2 := model.User{Role: "user", Status: 0}
	db.Create(&u1)
	db.Create(&u2)
	db.Create(&model.UserProfile{UserID: u1.ID, Nickname: "a"})
	db.Create(&model.UserProfile{UserID: u2.ID, Nickname: "b"})

	if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
		t.Fatal(err)
	}

	var p1, p2 model.UserProfile
	db.First(&p1, "user_id = ?", u1.ID)
	db.First(&p2, "user_id = ?", u2.ID)
	if p1.FollowingCount != 1 || p2.FollowerCount != 1 {
		t.Fatalf("counts not updated correctly: %d/%d", p1.FollowingCount, p2.FollowerCount)
	}
}
