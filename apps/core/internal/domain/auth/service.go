package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/token"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/email"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service exposes auth operations used by handlers.
type Service interface {
	Register(ctx context.Context, username, userEmail, password, baseURL string) (*model.User, error)
	Login(ctx context.Context, email, password, ipAddress string) (string, error)
	LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, avatar string) (string, error)
	VerifyEmail(ctx context.Context, token string) error
}

type service struct {
	db           *gorm.DB
	tokenService *token.Service
	emailClient  *email.SMTPClient
}

// NewService creates an auth service.
func NewService(db *gorm.DB, jwtCfg config.JWTConfig, smtpCfg config.SMTPConfig) Service {
	if db == nil {
		db = database.DB
	}
	var emailClient *email.SMTPClient
	if smtpCfg.Host != "" && smtpCfg.Port != 0 {
		emailClient = email.NewSMTPClient(smtpCfg)
	}
	return &service{
		db:           db,
		tokenService: token.New(jwtCfg),
		emailClient:  emailClient,
	}
}

func (s *service) Register(ctx context.Context, username, userEmail, password, baseURL string) (*model.User, error) {
	var existingAuth model.UserAuth
	if err := s.db.Where("identity_type = ? AND identifier = ?", "email", userEmail).First(&existingAuth).Error; err == nil {
		return nil, errors.New("user already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	token := uuid.New().String()

	var user model.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		user = model.User{Role: "user", Status: 2}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		auth := model.UserAuth{UserID: user.ID, IdentityType: "email", Identifier: userEmail}
		if err := auth.SetPassword(password); err != nil {
			return err
		}
		if err := tx.Create(&auth).Error; err != nil {
			return err
		}

		profile := model.UserProfile{UserID: user.ID, Nickname: username}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

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

	if s.emailClient == nil {
		logger.Warn(ctx, "SMTP not configured; verification email not sent",
			"event", "user_registered",
			"user_id", user.ID,
			"email", userEmail,
		)
		logger.Info(ctx, fmt.Sprintf("Verification link for %s: %s", userEmail, verifyURL),
			"event", "user_registered",
			"user_id", user.ID,
			"email", userEmail,
		)
	} else if err := s.emailClient.SendVerificationEmail(userEmail, verifyURL); err != nil {
		logger.Warn(ctx, "Failed to send verification email",
			"error", err.Error(),
			"user_id", user.ID,
			"email", userEmail,
		)
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

func (s *service) Login(ctx context.Context, email, password, ipAddress string) (string, error) {
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

	var user model.User
	if err := s.db.First(&user, auth.UserID).Error; err != nil {
		return "", err
	}

	if user.Status == 2 {
		return "", errors.New("please verify your email before logging in")
	}
	if user.Status == 1 {
		return "", errors.New("your account has been suspended")
	}

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

	_ = ipAddress
	return token, nil
}

func (s *service) LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, avatar string) (string, error) {
	var auth model.UserAuth
	var user model.User

	err := s.db.Where("identity_type = ? AND identifier = ?", provider, email).First(&auth).Error
	if err == nil {
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

	err = s.db.Transaction(func(tx *gorm.DB) error {
		user = model.User{Role: "user", Status: 0}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

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

		profile := model.UserProfile{UserID: user.ID, Nickname: name, AvatarURL: avatar}
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

func (s *service) VerifyEmail(ctx context.Context, token string) error {
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

	if err := s.db.Model(&model.User{}).Where("id = ?", verification.UserID).Update("status", 0).Error; err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

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
