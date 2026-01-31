package service

import (
	"context"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testJWTConfig = config.JWTConfig{
	Secret:     "test-secret-key-for-testing",
	ExpireHour: 24,
}

var testSMTPConfig = config.SMTPConfig{
	Host:     "localhost",
	Port:     25,
	Username: "",
	Password: "",
	From:     "test@example.com",
	UseTLS:   false,
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: nil, // Use default logger
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&model.User{}, &model.UserAuth{}, &model.UserProfile{}, &model.EmailVerification{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestRegister(t *testing.T) {
	db := setupTestDB(t)
	authService := NewAuthService(db, testJWTConfig, testSMTPConfig)

	ctx := context.Background()
	username := "testuser"
	email := "test@example.com"
	password := "password123"
	baseURL := "http://localhost:8080"

	// Test Case 1: Success
	user, err := authService.Register(ctx, username, email, password, baseURL)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("Expected user ID to be generated")
	}
	if user.Role != "user" {
		t.Errorf("Expected role 'user', got %s", user.Role)
	}
	if user.Status != 2 {
		t.Errorf("Expected status 2 (pending), got %d", user.Status)
	}

	// Verify that email verification record was created
	var verification model.EmailVerification
	if err := db.Where("user_id = ?", user.ID).First(&verification).Error; err != nil {
		t.Errorf("Expected email verification record to be created: %v", err)
	}

	// Test Case 2: Duplicate Email
	_, err = authService.Register(ctx, "otheruser", email, "pass", baseURL)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	authService := NewAuthService(db, testJWTConfig, testSMTPConfig)

	ctx := context.Background()
	email := "login@example.com"
	password := "securepass"
	username := "loginuser"
	baseURL := "http://localhost"

	// Create user
	user, err := authService.Register(ctx, username, email, password, baseURL)
	if err != nil {
		t.Fatalf("Failed to create user for login test: %v", err)
	}

	// Test Case 1: Login should fail for unverified user
	_, err = authService.Login(ctx, email, password, "127.0.0.1")
	if err == nil {
		t.Error("Expected error for unverified user, got nil")
	}

	// Verify the user's email first
	var verification model.EmailVerification
	if err := db.Where("user_id = ?", user.ID).First(&verification).Error; err != nil {
		t.Fatalf("Failed to find verification record: %v", err)
	}
	if err := authService.VerifyEmail(ctx, verification.Token); err != nil {
		t.Fatalf("Failed to verify email: %v", err)
	}

	// Test Case 2: Success after verification
	token, err := authService.Login(ctx, email, password, "127.0.0.1")
	if err != nil {
		t.Errorf("Login failed: %v", err)
	}
	if token == "" {
		t.Error("Expected JWT token, got empty string")
	}

	// Test Case 3: Wrong Password
	_, err = authService.Login(ctx, email, "wrongpass", "127.0.0.1")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}

	// Test Case 4: User Not Found
	_, err = authService.Login(ctx, "nonexistent@example.com", password, "127.0.0.1")
	if err == nil {
		t.Error("Expected error for user not found, got nil")
	}
}

func TestUserProfileHasCounts(t *testing.T) {
	db := setupTestDB(t)
	type Column struct {
		Name string
	}
	var cols []Column
	if err := db.Raw("PRAGMA table_info(user_profiles)").Scan(&cols).Error; err != nil {
		t.Fatalf("schema query failed: %v", err)
	}
	want := map[string]bool{
		"follower_count":  true,
		"following_count": true,
		"post_count":      true,
		"review_count":    true,
		"like_count":      true,
	}
	for _, c := range cols {
		if _, ok := want[c.Name]; ok {
			delete(want, c.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing columns: %v", want)
	}
}
