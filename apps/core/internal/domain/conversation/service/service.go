package service

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type ConversationService struct {
	db *gorm.DB
}

func NewConversationService(db *gorm.DB) *ConversationService {
	if db == nil {
		db = database.DB
	}
	return &ConversationService{db: db}
}
