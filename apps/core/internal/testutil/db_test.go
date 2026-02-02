package testutil

import (
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

func TestSetupTestDB(t *testing.T) {
	db := SetupTestDB(t)
	if db == nil {
		t.Fatal("expected db")
	}
	if !db.Migrator().HasTable(&model.User{}) {
		t.Fatal("expected user table to exist")
	}
}
