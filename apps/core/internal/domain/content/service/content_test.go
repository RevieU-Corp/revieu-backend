package service

import (
	"context"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
)

func TestContentServiceListPosts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := NewContentService(db)
	user := model.User{Role: "user", Status: 0}
	db.Create(&user)
	db.Create(&model.Post{UserID: user.ID, Content: "a"})
	posts, total, err := svc.ListUserPosts(context.Background(), user.ID, nil, 10)
	if err != nil || total != 1 || len(posts) != 1 {
		t.Fatalf("list posts failed: %v", err)
	}
}
