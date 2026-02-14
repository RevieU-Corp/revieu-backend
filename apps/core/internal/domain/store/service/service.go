package service

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type StoreService struct {
	db *gorm.DB
}

func NewStoreService(db *gorm.DB) *StoreService {
	if db == nil {
		db = database.DB
	}
	return &StoreService{db: db}
}
