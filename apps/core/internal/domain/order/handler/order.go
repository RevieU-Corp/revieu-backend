package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/order/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	if svc == nil {
		svc = service.NewOrderService(nil)
	}
	return &OrderHandler{svc: svc}
}

// CreateOrder godoc
// @Summary Create order
// @Description Creates a new order for the authenticated user
// @Tags order
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// ListOrders godoc
// @Summary List orders
// @Description Returns orders for the authenticated user
// @Tags order
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// OrderDetail godoc
// @Summary Get order detail
// @Description Returns an order by ID
// @Tags order
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /orders/{id} [get]
func (h *OrderHandler) Detail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}
