package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/email"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	db           *gorm.DB
	tokenService *TokenService
	emailClient  *email.SMTPClient
}

func NewAuthService(db *gorm.DB, jwtCfg config.JWTConfig, smtpCfg config.SMTPConfig) *AuthService {
	if db == nil {
		db = database.DB
	}
	return &AuthService{
		db:           db,
		tokenService: NewTokenService(jwtCfg),
		emailClient:  email.NewSMTPClient(smtpCfg),
	}
}

func (s *AuthService) Register(ctx context.Context, username, userEmail, password, baseURL string) (*model.User, error) {
	// Check if email already exists in user_auths
	var existingAuth model.UserAuth
	if err := s.db.Where("identity_type = ? AND identifier = ?", "email", userEmail).First(&existingAuth).Error; err == nil {
		return nil, errors.New("user already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Generate verification token
	token := uuid.New().String()

	// Create user, auth, profile, and verification in a transaction
	var user model.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create user with pending status
		user = model.User{
			Role:   "user",
			Status: 2, // pending - waiting for email verification
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Create user auth
		auth := model.UserAuth{
			UserID:       user.ID,
			IdentityType: "email",
			Identifier:   userEmail,
		}
		if err := auth.SetPassword(password); err != nil {
			return err
		}
		if err := tx.Create(&auth).Error; err != nil {
			return err
		}

		// Create user profile
		profile := model.UserProfile{
			UserID:   user.ID,
			Nickname: username,
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		// Create email verification record
		verification := model.EmailVerification{
			UserID:    user.ID,
			Email:     userEmail,
			Token:     token,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}
		if err := tx.Create(&verification).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)

	// Send verification email
	if err := s.emailClient.SendVerificationEmail(userEmail, verifyURL); err != nil {
		logger.Warn(ctx, "Failed to send verification email",
			"error", err.Error(),
			"user_id", user.ID,
			"email", userEmail,
		)
		// Still log the verification link for debugging
		logger.Info(ctx, fmt.Sprintf("Verification link for %s: %s", userEmail, verifyURL),
			"event", "user_registered",
			"user_id", user.ID,
			"email", userEmail,
		)
	} else {
		logger.Info(ctx, "Verification email sent",
			"event", "user_registered",
			"user_id", user.ID,
			"email", userEmail,
		)
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, ipAddress string) (string, error) {
	// Find user auth by email
	var auth model.UserAuth
	if err := s.db.Where("identity_type = ? AND identifier = ?", "email", email).First(&auth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	if !auth.CheckPassword(password) {
		return "", errors.New("invalid credentials")
	}

	// Get user
	var user model.User
	if err := s.db.First(&user, auth.UserID).Error; err != nil {
		return "", err
	}

	// Check user status
	if user.Status == 2 { // pending - email not verified
		return "", errors.New("please verify your email before logging in")
	}
	if user.Status == 1 { // banned
		return "", errors.New("your account has been suspended")
	}

	// Update login info
	now := time.Now().UTC()
	auth.LastLoginAt = &now
	if err := s.db.Save(&auth).Error; err != nil {
		logger.Warn(ctx, "Failed to update user login info",
			"user_id", user.ID,
			"error", err.Error(),
		)
	}

	token, err := s.tokenService.GenerateToken(&user, &auth)
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "User logged in successfully",
		"event", "user_login_success",
		"user_id", user.ID,
	)

	return token, nil
}

// LoginOrRegisterOAuthUser handles OAuth login/registration for providers like Google
func (s *AuthService) LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, avatar string) (string, error) {
	var auth model.UserAuth
	var user model.User

	// Try to find existing OAuth auth
	err := s.db.Where("identity_type = ? AND identifier = ?", provider, email).First(&auth).Error
	if err == nil {
		// User exists, update login time and return token
		if err := s.db.First(&user, auth.UserID).Error; err != nil {
			return "", err
		}

		now := time.Now().UTC()
		auth.LastLoginAt = &now
		if err := s.db.Save(&auth).Error; err != nil {
			logger.Warn(ctx, "Failed to update OAuth user login info",
				"user_id", user.ID,
				"error", err.Error(),
			)
		}

		token, err := s.tokenService.GenerateToken(&user, &auth)
		if err != nil {
			return "", err
		}

		logger.Info(ctx, "OAuth user logged in successfully",
			"event", "oauth_login_success",
			"user_id", user.ID,
			"provider", provider,
		)

		return token, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// User doesn't exist, create new user
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create user
		user = model.User{
			Role:   "user",
			Status: 0, // active
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Create user auth for OAuth
		now := time.Now().UTC()
		auth = model.UserAuth{
			UserID:       user.ID,
			IdentityType: provider,
			Identifier:   email,
			LastLoginAt:  &now,
		}
		if err := tx.Create(&auth).Error; err != nil {
			return err
		}

		// Create user profile
		profile := model.UserProfile{
			UserID:    user.ID,
			Nickname:  name,
			AvatarURL: avatar,
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	token, err := s.tokenService.GenerateToken(&user, &auth)
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "OAuth user registered and logged in successfully",
		"event", "oauth_register_success",
		"user_id", user.ID,
		"provider", provider,
	)

	return token, nil
}

// VerifyEmail verifies a user's email using the verification token
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	var verification model.EmailVerification
	if err := s.db.Where("token = ?", token).First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("invalid or expired verification token")
		}
		return err
	}

	if verification.IsExpired() {
		return errors.New("verification token has expired")
	}

	// Update user status to active
	if err := s.db.Model(&model.User{}).Where("id = ?", verification.UserID).Update("status", 0).Error; err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	// Delete the verification record
	if err := s.db.Delete(&verification).Error; err != nil {
		logger.Warn(ctx, "Failed to delete verification record",
			"error", err.Error(),
			"user_id", verification.UserID,
		)
	}

	logger.Info(ctx, "User email verified successfully",
		"event", "email_verified",
		"user_id", verification.UserID,
		"email", verification.Email,
	)

	return nil
}
