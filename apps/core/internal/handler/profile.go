package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	userService    *service.UserService
	followService  *service.FollowService
	contentService *service.ContentService
	db             *gorm.DB
}

func NewProfileHandler(userService *service.UserService, followService *service.FollowService, contentService *service.ContentService) *ProfileHandler {
	if userService == nil {
		userService = service.NewUserService(nil)
	}
	if followService == nil {
		followService = service.NewFollowService(nil)
	}
	if contentService == nil {
		contentService = service.NewContentService(nil)
	}
	return &ProfileHandler{
		userService:    userService,
		followService:  followService,
		contentService: contentService,
		db:             database.DB,
	}
}

func (h *ProfileHandler) GetPublicProfile(c *gin.Context) {
	targetID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	viewerID := c.GetInt64("user_id")
	var privacy model.UserPrivacy
	if err := h.db.WithContext(c.Request.Context()).First(&privacy, "user_id = ?", targetID).Error; err == nil {
		if !privacy.IsPublic && viewerID != targetID {
			c.JSON(http.StatusForbidden, gin.H{"error": "profile is private"})
			return
		}
	}

	var profile model.UserProfile
	if err := h.db.WithContext(c.Request.Context()).First(&profile, "user_id = ?", targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	isFollowing := false
	if viewerID != 0 && viewerID != targetID {
		var follow model.UserFollow
		if err := h.db.WithContext(c.Request.Context()).Where("follower_id = ? AND following_id = ?", viewerID, targetID).First(&follow).Error; err == nil {
			isFollowing = true
		}
	}

	c.JSON(http.StatusOK, dto.PublicProfileResponse{
		UserID:         profile.UserID,
		Nickname:       profile.Nickname,
		AvatarURL:      profile.AvatarURL,
		Intro:          profile.Intro,
		Location:       profile.Location,
		FollowerCount:  profile.FollowerCount,
		FollowingCount: profile.FollowingCount,
		PostCount:      profile.PostCount,
		ReviewCount:    profile.ReviewCount,
		LikeCount:      profile.LikeCount,
		IsFollowing:    isFollowing,
	})
}

func (h *ProfileHandler) ListUserPosts(c *gin.Context) {
	targetID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	cursor, limit := parseCursorLimit(c)
	posts, total, err := h.contentService.ListUserPosts(c.Request.Context(), targetID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	liked := h.likedIDs(c, "post")
	items := make([]dto.PostItem, 0, len(posts))
	for _, post := range posts {
		var merchant *dto.MerchantBrief
		if post.MerchantID != nil {
			merchant = h.loadMerchantBrief(*post.MerchantID)
		}
		items = append(items, dto.PostItem{
			ID:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			Images:    parseJSONStrings(post.Images),
			LikeCount: post.LikeCount,
			ViewCount: post.ViewCount,
			IsLiked:   liked[post.ID],
			Merchant:  merchant,
			Tags:      []string{},
			CreatedAt: post.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, dto.PostListResponse{Posts: items, Total: int(total), Cursor: nextCursor(posts)})
}

func (h *ProfileHandler) ListUserReviews(c *gin.Context) {
	targetID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	cursor, limit := parseCursorLimit(c)
	reviews, total, err := h.contentService.ListUserReviews(c.Request.Context(), targetID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	liked := h.likedIDs(c, "review")
	items := make([]dto.ReviewItem, 0, len(reviews))
	for _, review := range reviews {
		merchant := h.loadMerchantBrief(review.MerchantID)
		merchantValue := dto.MerchantBrief{}
		if merchant != nil {
			merchantValue = *merchant
		}
		items = append(items, dto.ReviewItem{
			ID:            review.ID,
			Rating:        review.Rating,
			RatingEnv:     review.RatingEnv,
			RatingService: review.RatingService,
			RatingValue:   review.RatingValue,
			Content:       review.Content,
			Images:        parseJSONStrings(review.Images),
			AvgCost:       review.AvgCost,
			LikeCount:     review.LikeCount,
			IsLiked:       liked[review.ID],
			Merchant:      merchantValue,
			Tags:          []string{},
			CreatedAt:     review.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, dto.ReviewListResponse{Reviews: items, Total: int(total), Cursor: nextCursor(reviews)})
}

func (h *ProfileHandler) FollowUser(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.followService.FollowUser(c.Request.Context(), userID, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ProfileHandler) UnfollowUser(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.followService.UnfollowUser(c.Request.Context(), userID, targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ProfileHandler) FollowMerchant(c *gin.Context) {
	userID := c.GetInt64("user_id")
	merchantID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}
	if err := h.followService.FollowMerchant(c.Request.Context(), userID, merchantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ProfileHandler) UnfollowMerchant(c *gin.Context) {
	userID := c.GetInt64("user_id")
	merchantID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}
	if err := h.followService.UnfollowMerchant(c.Request.Context(), userID, merchantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ProfileHandler) likedIDs(c *gin.Context, targetType string) map[int64]bool {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		return map[int64]bool{}
	}
	var likes []model.Like
	if err := h.db.WithContext(c.Request.Context()).Where("user_id = ? AND target_type = ?", userID, targetType).Find(&likes).Error; err != nil {
		return map[int64]bool{}
	}
	result := make(map[int64]bool, len(likes))
	for _, like := range likes {
		result[like.TargetID] = true
	}
	return result
}

func (h *ProfileHandler) loadMerchantBrief(id int64) *dto.MerchantBrief {
	var merchant model.Merchant
	if err := h.db.First(&merchant, id).Error; err != nil {
		return nil
	}
	return &dto.MerchantBrief{ID: merchant.ID, Name: merchant.Name, Category: merchant.Category}
}
