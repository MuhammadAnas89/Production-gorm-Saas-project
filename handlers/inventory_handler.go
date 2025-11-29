package handlers

import (
	"go-multi-tenant/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InventoryHandler struct {
	invService *services.InventoryService
}

func NewInventoryHandler(invService *services.InventoryService) *InventoryHandler {
	return &InventoryHandler{invService: invService}
}

// 1. Update Stock
func (h *InventoryHandler) UpdateStock(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.invService.UpdateStock(tenantDB, req.ProductID, req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock updated successfully"})
}

// 2. Get Low Stock Alerts
func (h *InventoryHandler) GetLowStockAlerts(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	thresholdStr := c.DefaultQuery("threshold", "10")
	threshold, _ := strconv.Atoi(thresholdStr)

	items, err := h.invService.GetLowStockAlerts(tenantDB, threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}
