package main

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/model"
	"gorm.io/gorm"
)

func runAutoMigrate(db *gorm.DB, enabled bool) error {
	if !enabled {
		return nil
	}
	return db.AutoMigrate(model.All()...)
}
