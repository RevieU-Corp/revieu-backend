package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
	"gorm.io/gorm"
)

// Tunables for the style-learning pipeline. Values are intentionally hard-coded for now —
// promote them to GeminiConfig if a deployment ever needs different cadence.
const (
	// DeriveTriggerN is how many newly submitted reviews accumulate before a derive job fires.
	DeriveTriggerN = 5
	// DeriveSampleK is how many of the user's most recent reviews are fed into the derive prompt.
	// We deliberately pull more than DeriveTriggerN so each derivation sees a wider voice window.
	DeriveSampleK = 10
	// deriveLockTTL bounds how long a soft lock blocks retries. If a job crashes silently the
	// lock self-expires after this window.
	deriveLockTTL = 5 * time.Minute
	// deriveJobTimeout caps a background derive run, including the Gemini call and DB writes.
	deriveJobTimeout = 60 * time.Second
)

// ReviewForStyle is the slim view of a review that the derive prompt needs.
type ReviewForStyle struct {
	ID      int64
	Content string
}

// StyleClient is the narrow Gemini surface used by the derive job. It is a separate
// interface from GeminiClient so existing polish-only tests do not need to implement it.
type StyleClient interface {
	DeriveStyle(ctx context.Context, reviews []ReviewForStyle) (dto.StyleProfile, []dto.SampleSnippet, error)
}

// StyleService owns the writing-style profile lifecycle: counter increments on review
// submission, async derivation when the threshold is hit, and lookups for the polish
// prompt. All methods are nil-safe so callers can treat a disabled AI feature as "no
// personalization" without conditional code.
type StyleService struct {
	db     *gorm.DB
	client StyleClient
}

// NewStyleService wires a StyleService. db falls back to the package-global database.DB
// when nil, matching the convention used by other domain services. client may be nil
// when the AI feature is disabled — every method on the service then becomes a no-op.
func NewStyleService(db *gorm.DB, client StyleClient) *StyleService {
	if db == nil {
		db = database.DB
	}
	return &StyleService{db: db, client: client}
}

// GetStyle returns the user's profile and samples, or (nil, nil, nil) when the user has
// no profile yet (cold start). Errors are returned but every call site is expected to
// treat them as "no profile" — the polish path must never fail because style lookup did.
func (s *StyleService) GetStyle(ctx context.Context, userID int64) (*dto.StyleProfile, []dto.SampleSnippet, error) {
	if s == nil || s.db == nil {
		return nil, nil, nil
	}

	var row model.UserWritingStyle
	err := s.db.WithContext(ctx).First(&row, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if row.LastDerivedAt == nil {
		// Row exists (counter has been bumped) but no successful derive has produced features yet.
		return nil, nil, nil
	}

	var profile dto.StyleProfile
	if row.Features != "" {
		if err := json.Unmarshal([]byte(row.Features), &profile); err != nil {
			return nil, nil, fmt.Errorf("style: decode features: %w", err)
		}
	}
	var samples []dto.SampleSnippet
	if row.Samples != "" {
		if err := json.Unmarshal([]byte(row.Samples), &samples); err != nil {
			return nil, nil, fmt.Errorf("style: decode samples: %w", err)
		}
	}
	return &profile, samples, nil
}

// OnReviewSubmitted increments the user's derive counter and, when the threshold is
// crossed, fires a background derive job. The hook is fire-and-forget: errors are
// logged, never returned, so the review-creation path is never blocked or rolled back
// by style accounting.
func (s *StyleService) OnReviewSubmitted(ctx context.Context, userID int64) {
	if s == nil || s.db == nil || s.client == nil {
		return
	}

	if err := s.bumpCounter(ctx, userID); err != nil {
		logger.Warn(ctx, "style: counter increment failed", "user_id", userID, "error", err.Error())
		return
	}

	row, err := s.fetchRow(ctx, userID)
	if err != nil {
		logger.Warn(ctx, "style: counter readback failed", "user_id", userID, "error", err.Error())
		return
	}
	if row.ReviewsSinceLastDerive < DeriveTriggerN {
		return
	}

	// Soft-lock acquisition via conditional UPDATE: only one goroutine wins, others bail.
	// The TTL guards against orphaned locks if the previous holder crashed.
	if !s.tryAcquireLock(ctx, userID) {
		return
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), deriveJobTimeout)
		defer cancel()
		if err := s.Derive(bgCtx, userID); err != nil {
			logger.Warn(bgCtx, "style: derive failed", "user_id", userID, "error", err.Error())
			s.releaseLock(bgCtx, userID)
		}
	}()
}

// Derive runs one synchronous derivation cycle: read the latest K reviews, call Gemini,
// persist the result. The caller is responsible for having acquired the soft lock; on
// success the lock is released and the counter is reset within the same UPDATE.
func (s *StyleService) Derive(ctx context.Context, userID int64) error {
	if s == nil || s.client == nil {
		return errors.New("style: service not configured")
	}

	var reviews []model.Review
	if err := s.db.WithContext(ctx).
		Select("id, content").
		Where("user_id = ? AND content <> ''", userID).
		Order("created_at DESC").
		Limit(DeriveSampleK).
		Find(&reviews).Error; err != nil {
		return fmt.Errorf("style: load reviews: %w", err)
	}
	if len(reviews) < DeriveTriggerN {
		// Edge case: counter says trigger fired but reviews were deleted between increment
		// and derive. Release the lock and bail without writing a profile.
		s.releaseLock(ctx, userID)
		return nil
	}

	input := make([]ReviewForStyle, 0, len(reviews))
	for _, r := range reviews {
		input = append(input, ReviewForStyle{ID: r.ID, Content: r.Content})
	}

	profile, snippets, err := s.client.DeriveStyle(ctx, input)
	if err != nil {
		return fmt.Errorf("style: gemini derive: %w", err)
	}

	featuresJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("style: encode features: %w", err)
	}
	samplesJSON, err := json.Marshal(snippets)
	if err != nil {
		return fmt.Errorf("style: encode samples: %w", err)
	}

	now := time.Now().UTC()
	res := s.db.WithContext(ctx).Model(&model.UserWritingStyle{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"features":                  string(featuresJSON),
			"samples":                   string(samplesJSON),
			"reviews_since_last_derive": 0,
			"derived_from_review_count": len(reviews),
			"last_derived_at":           now,
			"last_derive_started_at":    nil,
		})
	if res.Error != nil {
		return fmt.Errorf("style: persist profile: %w", res.Error)
	}
	return nil
}

// bumpCounter ensures the row exists and increments the per-derive counter atomically.
// The Postgres UPSERT keeps inserts and increments in a single round trip and avoids the
// classic read-modify-write race when two reviews land at once.
func (s *StyleService) bumpCounter(ctx context.Context, userID int64) error {
	const sql = `
		INSERT INTO user_writing_styles (user_id, reviews_since_last_derive, updated_at)
		VALUES (?, 1, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET reviews_since_last_derive = user_writing_styles.reviews_since_last_derive + 1,
		    updated_at = NOW()
	`
	return s.db.WithContext(ctx).Exec(sql, userID).Error
}

func (s *StyleService) fetchRow(ctx context.Context, userID int64) (*model.UserWritingStyle, error) {
	var row model.UserWritingStyle
	if err := s.db.WithContext(ctx).First(&row, userID).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// tryAcquireLock returns true when this caller obtained the derive lock for the user.
// A previous holder is considered abandoned once its timestamp is older than deriveLockTTL.
func (s *StyleService) tryAcquireLock(ctx context.Context, userID int64) bool {
	now := time.Now().UTC()
	cutoff := now.Add(-deriveLockTTL)
	res := s.db.WithContext(ctx).Model(&model.UserWritingStyle{}).
		Where("user_id = ? AND (last_derive_started_at IS NULL OR last_derive_started_at < ?)", userID, cutoff).
		Update("last_derive_started_at", now)
	return res.Error == nil && res.RowsAffected > 0
}

func (s *StyleService) releaseLock(ctx context.Context, userID int64) {
	_ = s.db.WithContext(ctx).Model(&model.UserWritingStyle{}).
		Where("user_id = ?", userID).
		Update("last_derive_started_at", nil).Error
}
