package service

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	if db == nil {
		db = database.DB
	}
	return &CategoryService{db: db}
}
