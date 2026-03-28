package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/notification/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNotificationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Notification{}); err != nil {
		t.Fatalf("failed to migrate notification test db: %v", err)
	}

	return db
}

func seedNotificationFixture(t *testing.T, db *gorm.DB) {
	t.Helper()

	user := model.User{ID: 601, Role: "merchant", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	notifications := []model.Notification{
		{
			ID:        8001,
			UserID:    user.ID,
			Type:      "verification",
			Title:     "Verification submitted",
			Content:   "We received your business documents.",
			IsRead:    false,
			CreatedAt: time.Now().Add(-time.Hour),
		},
	}
	if err := db.Create(&notifications).Error; err != nil {
		t.Fatalf("failed to create notifications: %v", err)
	}
}

func TestNotificationHandlerListReturnsNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupNotificationTestDB(t)
	seedNotificationFixture(t, db)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/notifications", nil)
	c.Set("user_id", int64(601))

	h := NewNotificationHandler(service.NewNotificationService(db))
	h.List(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("not implemented")) {
		t.Fatalf("expected notification list payload, got %s", recorder.Body.String())
	}

	var response struct {
		Data []struct {
			ID      int64  `json:"id"`
			Title   string `json:"title"`
			Content string `json:"content"`
			IsRead  bool   `json:"is_read"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Data) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(response.Data))
	}
	if response.Data[0].Title != "Verification submitted" {
		t.Fatalf("expected verification notification title, got %q", response.Data[0].Title)
	}
}

func TestNotificationHandlerMarkReadUpdatesNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupNotificationTestDB(t)
	seedNotificationFixture(t, db)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: "8001"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/notifications/8001/read", nil)
	c.Set("user_id", int64(601))

	h := NewNotificationHandler(service.NewNotificationService(db))
	h.MarkRead(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("not implemented")) {
		t.Fatalf("expected mark-read payload, got %s", recorder.Body.String())
	}

	var notification model.Notification
	if err := db.First(&notification, 8001).Error; err != nil {
		t.Fatalf("failed to reload notification: %v", err)
	}
	if !notification.IsRead {
		t.Fatalf("expected notification to be marked read")
	}
}
