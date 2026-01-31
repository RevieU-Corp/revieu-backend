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
