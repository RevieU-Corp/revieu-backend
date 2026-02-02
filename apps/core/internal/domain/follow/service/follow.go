package service

import (
	"context"
	"errors"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type FollowService struct {
	db *gorm.DB
}

func NewFollowService(db *gorm.DB) *FollowService {
	if db == nil {
		db = database.DB
	}
	return &FollowService{db: db}
}

func (s *FollowService) FollowUser(ctx context.Context, followerID, followingID int64) error {
	if followerID == followingID {
		return errors.New("cannot follow self")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.UserFollow
		result := tx.Where("follower_id = ? AND following_id = ?", followerID, followingID).Limit(1).Find(&existing)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			return nil
		}
		follow := model.UserFollow{FollowerID: followerID, FollowingID: followingID}
		if err := tx.Create(&follow).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count + 1")).Error; err != nil {
			return err
		}
		return tx.Model(&model.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
	})
}

func (s *FollowService) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.UserFollow{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		if err := tx.Model(&model.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("GREATEST(following_count - 1, 0)")).Error; err != nil {
			return err
		}
		return tx.Model(&model.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)")).Error
	})
}

func (s *FollowService) FollowMerchant(ctx context.Context, userID, merchantID int64) error {
	follow := model.MerchantFollow{UserID: userID, MerchantID: merchantID}
	return s.db.WithContext(ctx).FirstOrCreate(&follow, follow).Error
}

func (s *FollowService) UnfollowMerchant(ctx context.Context, userID, merchantID int64) error {
	return s.db.WithContext(ctx).Where("user_id = ? AND merchant_id = ?", userID, merchantID).Delete(&model.MerchantFollow{}).Error
}
