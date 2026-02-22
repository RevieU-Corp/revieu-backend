package service

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type VerificationService struct {
	db *gorm.DB
}

func NewVerificationService(db *gorm.DB) *VerificationService {
	if db == nil {
		db = database.DB
	}
	return &VerificationService{db: db}
}
