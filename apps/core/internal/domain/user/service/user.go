package service

import (
	"context"
	"errors"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	if db == nil {
		db = database.DB
	}
	return &UserService{db: db}
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (dto.ProfileResponse, error) {
	var profile model.UserProfile
	if err := s.db.WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err != nil {
		return dto.ProfileResponse{}, err
	}
	return dto.ProfileResponse{
		UserID:    profile.UserID,
		Nickname:  profile.Nickname,
		AvatarURL: profile.AvatarURL,
		Intro:     profile.Intro,
		Location:  profile.Location,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) error {
	updates := map[string]any{}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	if req.Intro != nil {
		updates["intro"] = *req.Intro
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (s *UserService) GetPrivacy(ctx context.Context, userID int64) (dto.PrivacySettings, error) {
	var privacy model.UserPrivacy
	if err := s.db.WithContext(ctx).FirstOrCreate(&privacy, model.UserPrivacy{UserID: userID}).Error; err != nil {
		return dto.PrivacySettings{}, err
	}
	return dto.PrivacySettings{IsPublic: privacy.IsPublic}, nil
}

func (s *UserService) UpdatePrivacy(ctx context.Context, userID int64, req dto.PrivacySettings) error {
	return s.db.WithContext(ctx).Save(&model.UserPrivacy{UserID: userID, IsPublic: req.IsPublic}).Error
}

func (s *UserService) GetNotifications(ctx context.Context, userID int64) (dto.NotificationSettings, error) {
	var notif model.UserNotification
	if err := s.db.WithContext(ctx).FirstOrCreate(&notif, model.UserNotification{UserID: userID}).Error; err != nil {
		return dto.NotificationSettings{}, err
	}
	return dto.NotificationSettings{PushEnabled: notif.PushEnabled, EmailEnabled: notif.EmailEnabled}, nil
}

func (s *UserService) UpdateNotifications(ctx context.Context, userID int64, req dto.NotificationSettings) error {
	return s.db.WithContext(ctx).Save(&model.UserNotification{UserID: userID, PushEnabled: req.PushEnabled, EmailEnabled: req.EmailEnabled}).Error
}

func (s *UserService) ListAddresses(ctx context.Context, userID int64) ([]model.UserAddress, error) {
	var items []model.UserAddress
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default desc, id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *UserService) CreateAddress(ctx context.Context, userID int64, req dto.CreateAddressRequest) (model.UserAddress, error) {
	var addr model.UserAddress
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
			return err
		}
		if count >= 20 {
			return errors.New("address limit reached")
		}
		addr = model.UserAddress{
			UserID:     userID,
			Name:       req.Name,
			Phone:      req.Phone,
			Province:   req.Province,
			City:       req.City,
			District:   req.District,
			Address:    req.Address,
			PostalCode: req.PostalCode,
			IsDefault:  req.IsDefault || count == 0,
		}
		if addr.IsDefault {
			if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&addr).Error
	})
	return addr, err
}

func (s *UserService) UpdateAddress(ctx context.Context, userID, addressID int64, req dto.UpdateAddressRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{}
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Phone != nil {
			updates["phone"] = *req.Phone
		}
		if req.Province != nil {
			updates["province"] = *req.Province
		}
		if req.City != nil {
			updates["city"] = *req.City
		}
		if req.District != nil {
			updates["district"] = *req.District
		}
		if req.Address != nil {
			updates["address"] = *req.Address
		}
		if req.PostalCode != nil {
			updates["postal_code"] = *req.PostalCode
		}
		if req.IsDefault != nil {
			updates["is_default"] = *req.IsDefault
		}
		if err := tx.Model(&model.UserAddress{}).Where("id = ? AND user_id = ?", addressID, userID).Updates(updates).Error; err != nil {
			return err
		}
		if req.IsDefault != nil && *req.IsDefault {
			return tx.Model(&model.UserAddress{}).Where("user_id = ? AND id <> ?", userID, addressID).Update("is_default", false).Error
		}
		return nil
	})
}

func (s *UserService) DeleteAddress(ctx context.Context, userID, addressID int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var addr model.UserAddress
		if err := tx.First(&addr, "id = ? AND user_id = ?", addressID, userID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&addr).Error; err != nil {
			return err
		}
		if addr.IsDefault {
			var replacement model.UserAddress
			if err := tx.Where("user_id = ?", userID).Order("id asc").First(&replacement).Error; err == nil {
				return tx.Model(&replacement).Update("is_default", true).Error
			}
		}
		return nil
	})
}

func (s *UserService) SetDefaultAddress(ctx context.Context, userID, addressID int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&model.UserAddress{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error
	})
}

func (s *UserService) RequestAccountDeletion(ctx context.Context, userID int64, reason string) error {
	deletion := model.AccountDeletion{UserID: userID, Reason: reason, ScheduledAt: time.Now().UTC().Add(7 * 24 * time.Hour)}
	return s.db.WithContext(ctx).Save(&deletion).Error
}
