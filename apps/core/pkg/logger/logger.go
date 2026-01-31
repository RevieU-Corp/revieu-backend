package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

var Log *slog.Logger

type Config struct {
	Service string
	Version string
	Level   string
}

func Init(cfg Config) {
	level := slog.LevelInfo
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
				// RFC3339 format is default
			}
			if a.Key == slog.LevelKey {
				a.Key = "level"
			}
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler).With(
		slog.String("service", cfg.Service),
		slog.String("version", cfg.Version),
	)
}

// Helper functions for easy logging
func Info(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelWarn, msg, args...)
}

func log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if Log == nil {
		// Fallback if not initialized
		handler := slog.NewJSONHandler(os.Stdout, nil)
		Log = slog.New(handler)
	}

	// Extract context values if available
	if ctx != nil {
		if traceID := ctx.Value("trace_id"); traceID != nil {
			args = append(args, slog.Any("trace_id", traceID))
		}
		if spanID := ctx.Value("span_id"); spanID != nil {
			args = append(args, slog.Any("span_id", spanID))
		}
		if userID := ctx.Value("user_id"); userID != nil {
			args = append(args, slog.Any("user_id", userID))
		}
		if clientIP := ctx.Value("client_ip"); clientIP != nil {
			args = append(args, slog.Any("client_ip", clientIP))
		}
	}
	
	// Add stack trace for errors
	if level == slog.LevelError {
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, false)
		args = append(args, slog.String("stack_trace", string(buf[:n])))
	}

	Log.Log(ctx, level, msg, args...)
}

