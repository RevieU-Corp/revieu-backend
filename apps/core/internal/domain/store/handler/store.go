package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/store/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/store/service"
	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	svc *service.StoreService
}

func NewStoreHandler(svc *service.StoreService) *StoreHandler {
	if svc == nil {
		svc = service.NewStoreService(nil)
	}
	return &StoreHandler{svc: svc}
}

// ListStores godoc
// @Summary List stores
// @Description Returns a list of stores
// @Tags store
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /stores [get]
func (h *StoreHandler) List(c *gin.Context) {
	stores, err := h.svc.ListPublished(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list stores"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stores})
}

// StoreDetail godoc
// @Summary Get store detail
// @Description Returns a store by ID
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /stores/{id} [get]
func (h *StoreHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	store, err := h.svc.DetailPublished(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load store"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": store})
}

// StoreReviews godoc
// @Summary Get store reviews
// @Description Returns reviews for a store
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /stores/{id}/reviews [get]
func (h *StoreHandler) Reviews(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	reviews, err := h.svc.ReviewsPublished(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load reviews"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reviews})
}

// StoreHours godoc
// @Summary Get store hours
// @Description Returns operating hours for a store
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /stores/{id}/hours [get]
func (h *StoreHandler) Hours(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	hours, err := h.svc.HoursPublished(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load hours"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": hours})
}

// ListMerchantStores godoc
// @Summary List current merchant stores
// @Description Returns stores owned by the authenticated merchant user
// @Tags store
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /merchant/stores [get]
func (h *StoreHandler) ListMine(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	stores, err := h.svc.ListMine(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list stores"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stores})
}

// CreateStore godoc
// @Summary Create a store
// @Description Creates a new store for the authenticated merchant
// @Tags store
// @Accept json
// @Produce json
// @Param request body dto.CreateStoreRequest false "Create store request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /merchant/stores [post]
func (h *StoreHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	store, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create store"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": store})
}

// UpdateStore godoc
// @Summary Update a store
// @Description Updates a store for the authenticated merchant
// @Tags store
// @Accept json
// @Produce json
// @Param id path int true "Store ID"
// @Param request body dto.UpdateStoreRequest false "Update store request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /merchant/stores/{id} [patch]
func (h *StoreHandler) Update(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	storeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	var req dto.UpdateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	store, err := h.svc.Update(c.Request.Context(), userID, storeID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStoreNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case errors.Is(err, service.ErrStoreForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update store"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": store})
}

// ActivateStore godoc
// @Summary Activate a store
// @Description Marks a merchant-owned store as published
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /merchant/stores/{id}/activate [post]
func (h *StoreHandler) Activate(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	storeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	if err := h.svc.Activate(c.Request.Context(), userID, storeID); err != nil {
		switch {
		case errors.Is(err, service.ErrStoreNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case errors.Is(err, service.ErrStoreForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate store"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeactivateStore godoc
// @Summary Deactivate a store
// @Description Marks a merchant-owned store as hidden
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /merchant/stores/{id}/deactivate [post]
func (h *StoreHandler) Deactivate(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	storeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	if err := h.svc.Deactivate(c.Request.Context(), userID, storeID); err != nil {
		switch {
		case errors.Is(err, service.ErrStoreNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case errors.Is(err, service.ErrStoreForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate store"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
