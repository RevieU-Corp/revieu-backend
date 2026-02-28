package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/store/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSeedRemoteStores100(t *testing.T) {
	if os.Getenv("SEED_REMOTE_STORES") != "1" {
		t.Skip("set SEED_REMOTE_STORES=1 to run")
	}

	host := os.Getenv("DB_HOST")
	password := os.Getenv("DB_PASSWORD")
	if host == "" || password == "" {
		t.Fatalf("DB_HOST/DB_PASSWORD required")
	}

	dsn := fmt.Sprintf("host=%s port=5432 user=postgres password=%s dbname=revieu sslmode=disable", host, password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}

	svc := NewStoreService(db)
	prefix := fmt.Sprintf("seed-%d", time.Now().Unix())

	var firstStoreID int64
	var lastStoreID int64
	for i := 1; i <= 100; i++ {
		user := model.User{Role: "user", Status: 0}
		if err := db.WithContext(context.Background()).Create(&user).Error; err != nil {
			t.Fatalf("create user %d failed: %v", i, err)
		}

		merchant := model.Merchant{
			UserID: &user.ID,
			Name:   fmt.Sprintf("%s-merchant-%03d", prefix, i),
		}
		if err := db.WithContext(context.Background()).Create(&merchant).Error; err != nil {
			t.Fatalf("create merchant %d failed: %v", i, err)
		}

		store, err := svc.Create(context.Background(), user.ID, dto.CreateStoreRequest{
			Name:    fmt.Sprintf("%s-store-%03d", prefix, i),
			City:    "Austin",
			State:   "TX",
			Country: "US",
			Phone:   fmt.Sprintf("+1-512-000-%04d", i),
			Images:  []string{fmt.Sprintf("https://example.com/%s/%03d.jpg", prefix, i)},
		})
		if err != nil {
			t.Fatalf("create store %d failed: %v", i, err)
		}

		if i == 1 {
			firstStoreID = store.ID
		}
		lastStoreID = store.ID
	}

	t.Logf("seed completed: prefix=%s count=100 first_store_id=%d last_store_id=%d", prefix, firstStoreID, lastStoreID)
}
