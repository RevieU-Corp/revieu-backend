package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService        *service.UserService
	followService      *service.FollowService
	contentService     *service.ContentService
	interactionService *service.InteractionService
	db                 *gorm.DB
}

func NewUserHandler(userService *service.UserService, followService *service.FollowService, contentService *service.ContentService, interactionService *service.InteractionService) *UserHandler {
	if userService == nil {
		userService = service.NewUserService(nil)
	}
	if followService == nil {
		followService = service.NewFollowService(nil)
	}
	if contentService == nil {
		contentService = service.NewContentService(nil)
	}
	if interactionService == nil {
		interactionService = service.NewInteractionService(nil)
	}
	return &UserHandler{
		userService:        userService,
		followService:      followService,
		contentService:     contentService,
		interactionService: interactionService,
		db:                 database.DB,
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userService.UpdateProfile(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) GetPrivacy(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := h.userService.GetPrivacy(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdatePrivacy(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req dto.PrivacySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userService.UpdatePrivacy(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) GetNotifications(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := h.userService.GetNotifications(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateNotifications(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req dto.NotificationSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userService.UpdateNotifications(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) ListAddresses(c *gin.Context) {
	userID := c.GetInt64("user_id")
	items, err := h.userService.ListAddresses(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := dto.AddressListResponse{Addresses: make([]dto.AddressItem, 0, len(items))}
	for _, item := range items {
		resp.Addresses = append(resp.Addresses, dto.AddressItem{
			ID:         item.ID,
			Name:       item.Name,
			Phone:      item.Phone,
			Province:   item.Province,
			City:       item.City,
			District:   item.District,
			Address:    item.Address,
			PostalCode: item.PostalCode,
			IsDefault:  item.IsDefault,
		})
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) CreateAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req dto.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	addr, err := h.userService.CreateAddress(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.AddressItem{
		ID:         addr.ID,
		Name:       addr.Name,
		Phone:      addr.Phone,
		Province:   addr.Province,
		City:       addr.City,
		District:   addr.District,
		Address:    addr.Address,
		PostalCode: addr.PostalCode,
		IsDefault:  addr.IsDefault,
	})
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	addressID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}
	var req dto.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userService.UpdateAddress(c.Request.Context(), userID, addressID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	addressID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}
	if err := h.userService.DeleteAddress(c.Request.Context(), userID, addressID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) SetDefaultAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	addressID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}
	if err := h.userService.SetDefaultAddress(c.Request.Context(), userID, addressID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) ListMyPosts(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cursor, limit := parseCursorLimit(c)
	posts, total, err := h.contentService.ListUserPosts(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]dto.PostItem, 0, len(posts))
	liked := h.likedIDs(c, "post")
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
	resp := dto.PostListResponse{Posts: items, Total: int(total), Cursor: nextCursor(posts)}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) ListMyReviews(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cursor, limit := parseCursorLimit(c)
	reviews, total, err := h.contentService.ListUserReviews(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]dto.ReviewItem, 0, len(reviews))
	liked := h.likedIDs(c, "review")
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
	resp := dto.ReviewListResponse{Reviews: items, Total: int(total), Cursor: nextCursor(reviews)}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) ListMyFavorites(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cursor, limit := parseCursorLimit(c)
	targetType := c.Query("type")
	items, total, err := h.contentService.ListFavorites(c.Request.Context(), userID, targetType, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	respItems := make([]dto.FavoriteItem, 0, len(items))
	for _, item := range items {
		fav := dto.FavoriteItem{
			ID:         item.ID,
			TargetType: item.TargetType,
			TargetID:   item.TargetID,
			CreatedAt:  item.CreatedAt,
		}
		switch item.TargetType {
		case "post":
			var post model.Post
			if err := h.db.WithContext(c.Request.Context()).First(&post, item.TargetID).Error; err == nil {
				fav.Post = &dto.PostItem{
					ID:        post.ID,
					Title:     post.Title,
					Content:   post.Content,
					Images:    parseJSONStrings(post.Images),
					LikeCount: post.LikeCount,
					ViewCount: post.ViewCount,
					IsLiked:   false,
					Merchant:  nil,
					Tags:      []string{},
					CreatedAt: post.CreatedAt,
				}
			}
		case "review":
			var review model.Review
			if err := h.db.WithContext(c.Request.Context()).First(&review, item.TargetID).Error; err == nil {
				fav.Review = &dto.ReviewItem{
					ID:            review.ID,
					Rating:        review.Rating,
					RatingEnv:     review.RatingEnv,
					RatingService: review.RatingService,
					RatingValue:   review.RatingValue,
					Content:       review.Content,
					Images:        parseJSONStrings(review.Images),
					AvgCost:       review.AvgCost,
					LikeCount:     review.LikeCount,
					IsLiked:       false,
					Merchant:      dto.MerchantBrief{},
					Tags:          []string{},
					CreatedAt:     review.CreatedAt,
				}
			}
		case "merchant":
			if merchant := h.loadMerchantBrief(item.TargetID); merchant != nil {
				fav.Merchant = merchant
			}
		}
		respItems = append(respItems, fav)
	}
	resp := dto.FavoriteListResponse{Items: respItems, Total: int(total), Cursor: nextCursor(items)}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) ListMyLikes(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cursor, limit := parseCursorLimit(c)
	items, total, err := h.contentService.ListLikes(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	respItems := make([]gin.H, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, gin.H{
			"id":          item.ID,
			"target_type": item.TargetType,
			"target_id":   item.TargetID,
			"created_at":  item.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"items":  respItems,
		"total":  total,
		"cursor": nextCursor(items),
	})
}

func (h *UserHandler) ListFollowingUsers(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cursor, limit := parseCursorLimit(c)
	var follows []model.UserFollow
	q := h.db.WithContext(c.Request.Context()).Where("follower_id = ?", userID).Order("id desc")
	if cursor != nil {
		q = q.Where("id < ?", *cursor)
	}
	if err := q.Limit(limit).Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ids := make([]int64, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.FollowingID)
	}
	profiles := h.loadUserBriefs(ids)
	c.JSON(http.StatusOK, dto.FollowingUsersResponse{Users: profiles, Total: len(profiles)})
}

func (h *UserHandler) ListFollowingMerchants(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cursor, limit := parseCursorLimit(c)
	var follows []model.MerchantFollow
	q := h.db.WithContext(c.Request.Context()).Where("user_id = ?", userID).Order("id desc")
	if cursor != nil {
		q = q.Where("id < ?", *cursor)
	}
	if err := q.Limit(limit).Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	merchants := make([]dto.MerchantBrief, 0, len(follows))
	for _, f := range follows {
		if merchant := h.loadMerchantBrief(f.MerchantID); merchant != nil {
			merchants = append(merchants, *merchant)
		}
	}
	c.JSON(http.StatusOK, gin.H{"merchants": merchants, "total": len(merchants)})
}

func (h *UserHandler) ListFollowers(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cursor, limit := parseCursorLimit(c)
	var follows []model.UserFollow
	q := h.db.WithContext(c.Request.Context()).Where("following_id = ?", userID).Order("id desc")
	if cursor != nil {
		q = q.Where("id < ?", *cursor)
	}
	if err := q.Limit(limit).Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ids := make([]int64, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.FollowerID)
	}
	profiles := h.loadUserBriefs(ids)
	c.JSON(http.StatusOK, dto.FollowersResponse{Users: profiles, Total: len(profiles)})
}

func (h *UserHandler) RequestAccountExport(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"status": "export queued"})
}

func (h *UserHandler) RequestAccountDeletion(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.userService.RequestAccountDeletion(c.Request.Context(), userID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "deletion_scheduled"})
}

func parseIDParam(c *gin.Context, name string) (int64, error) {
	return strconv.ParseInt(c.Param(name), 10, 64)
}

func parseCursorLimit(c *gin.Context) (*int64, int) {
	limit := 20
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	var cursor *int64
	if v := c.Query("cursor"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			cursor = &parsed
		}
	}
	return cursor, limit
}

func parseJSONStrings(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	return values
}

func nextCursor[T any](items []T) *int64 {
	return nil
}

func (h *UserHandler) likedIDs(c *gin.Context, targetType string) map[int64]bool {
	userID := c.GetInt64("user_id")
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

func (h *UserHandler) loadMerchantBrief(id int64) *dto.MerchantBrief {
	var merchant model.Merchant
	if err := h.db.First(&merchant, id).Error; err != nil {
		return nil
	}
	return &dto.MerchantBrief{ID: merchant.ID, Name: merchant.Name, Category: merchant.Category}
}

func (h *UserHandler) loadUserBriefs(ids []int64) []dto.UserBrief {
	if len(ids) == 0 {
		return []dto.UserBrief{}
	}
	var profiles []model.UserProfile
	if err := h.db.Where("user_id IN ?", ids).Find(&profiles).Error; err != nil {
		return []dto.UserBrief{}
	}
	result := make([]dto.UserBrief, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, dto.UserBrief{UserID: p.UserID, Nickname: p.Nickname, AvatarURL: p.AvatarURL, Intro: p.Intro})
	}
	return result
}
