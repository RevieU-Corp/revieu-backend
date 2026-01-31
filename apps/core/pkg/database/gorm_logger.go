package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	LogLevel      gormlogger.LogLevel
	SlowThreshold time.Duration
}

func NewGormLogger() *GormLogger {
	return &GormLogger{
		LogLevel:      gormlogger.Info,
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		logger.Info(ctx, fmt.Sprintf(msg, data...), "source", "gorm")
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		logger.Warn(ctx, fmt.Sprintf(msg, data...), "source", "gorm")
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		logger.Error(ctx, fmt.Sprintf(msg, data...), "source", "gorm")
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []interface{}{
		"source", "gorm",
		"duration_ms", float64(elapsed.Nanoseconds()) / 1e6,
		"rows", rows,
		"sql", sql,
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		fields = append(fields, "error", err.Error())
		logger.Error(ctx, "Database Query Failed", fields...)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		fields = append(fields, "slow_query", true)
		logger.Warn(ctx, "Slow SQL Query", fields...)
	case l.LogLevel >= gormlogger.Info:
		logger.Info(ctx, "SQL Query", fields...)
	}
}
