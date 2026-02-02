package service

import (
	"context"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
)

func TestInteractionServiceLike(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := NewInteractionService(db)
	u := model.User{Role: "user", Status: 0}
	db.Create(&u)
	if err := svc.Like(context.Background(), u.ID, "post", 123); err != nil {
		t.Fatal(err)
	}
	if err := svc.Unlike(context.Background(), u.ID, "post", 123); err != nil {
		t.Fatal(err)
	}
}
