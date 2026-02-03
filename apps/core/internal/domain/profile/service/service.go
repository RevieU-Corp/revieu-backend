package service

import (
	"context"
	"fmt"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type Service interface {
	GetPublicProfile(ctx context.Context, userID int64) (*PublicProfileResponse, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	if db == nil {
		db = database.DB
	}
	return &service{db: db}
}

type PublicProfileResponse struct {
	UserID         int64  `json:"user_id"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	Intro          string `json:"intro"`
	Location       string `json:"location"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	PostCount      int    `json:"post_count"`
	ReviewCount    int    `json:"review_count"`
	LikeCount      int    `json:"like_count"`
	IsFollowing    bool   `json:"is_following"`
}

func (s *service) GetPublicProfile(ctx context.Context, userID int64) (*PublicProfileResponse, error) {
	var profile model.UserProfile
	if err := s.db.WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &PublicProfileResponse{
		UserID:         profile.UserID,
		Nickname:       profile.Nickname,
		AvatarURL:      profile.AvatarURL,
		Intro:          profile.Intro,
		Location:       profile.Location,
		FollowerCount:  profile.FollowerCount,
		FollowingCount: profile.FollowingCount,
		PostCount:      profile.PostCount,
		ReviewCount:    profile.ReviewCount,
		LikeCount:      profile.LikeCount,
		IsFollowing:    false,
	}, nil
}
