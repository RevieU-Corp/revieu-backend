package main

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func tableExists(t *testing.T, db *gorm.DB, table string) bool {
	t.Helper()

	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?", "table", table).Scan(&count).Error; err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	return count > 0
}

func TestRunAutoMigrate_Disabled_NoSchemaChanges(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := runAutoMigrate(db, false); err != nil {
		t.Fatalf("runAutoMigrate returned error: %v", err)
	}

	if tableExists(t, db, "users") {
		t.Fatalf("expected users table to not exist when automigrate is disabled")
	}
}

func TestRunAutoMigrate_Enabled_CreatesCoreTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := runAutoMigrate(db, true); err != nil {
		t.Fatalf("runAutoMigrate returned error: %v", err)
	}

	if !tableExists(t, db, "users") {
		t.Fatalf("expected users table to exist when automigrate is enabled")
	}
	if !tableExists(t, db, "user_auths") {
		t.Fatalf("expected user_auths table to exist when automigrate is enabled")
	}
}
