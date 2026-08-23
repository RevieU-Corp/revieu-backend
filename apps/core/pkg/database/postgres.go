package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// dsnQuote escapes a libpq keyword/value connection-string parameter so
// values containing whitespace, single quotes, or backslashes (e.g. a
// generated password with a space in it) don't corrupt the parsing of
// subsequent key=value pairs.
func dsnQuote(value string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(value)
	return "'" + escaped + "'"
}

func Connect(cfg config.DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		dsnQuote(cfg.Host), dsnQuote(cfg.Username), dsnQuote(cfg.Password), dsnQuote(cfg.Database), cfg.Port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}
