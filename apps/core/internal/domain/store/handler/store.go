package handler

import (
	"net/http"

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
// @Router /stores [get]
func (h *StoreHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// StoreDetail godoc
// @Summary Get store detail
// @Description Returns a store by ID
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /stores/{id} [get]
func (h *StoreHandler) Detail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// StoreReviews godoc
// @Summary Get store reviews
// @Description Returns reviews for a store
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]interface{}
// @Router /stores/{id}/reviews [get]
func (h *StoreHandler) Reviews(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// StoreHours godoc
// @Summary Get store hours
// @Description Returns operating hours for a store
// @Tags store
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]interface{}
// @Router /stores/{id}/hours [get]
func (h *StoreHandler) Hours(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// CreateStore godoc
// @Summary Create a store
// @Description Creates a new store for the authenticated merchant
// @Tags store
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /merchant/stores [post]
func (h *StoreHandler) Create(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// UpdateStore godoc
// @Summary Update a store
// @Description Updates a store for the authenticated merchant
// @Tags store
// @Accept json
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /merchant/stores/{id} [patch]
func (h *StoreHandler) Update(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}
