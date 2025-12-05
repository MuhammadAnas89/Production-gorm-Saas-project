package handlers

import (
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PurchaseHandler struct {
	service *services.PurchaseService
}

func NewPurchaseHandler(service *services.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{service: service}
}

func (h *PurchaseHandler) Create(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	tenantID := c.MustGet("tenantID").(uint)
	userID := c.MustGet("userID").(uint)

	var req models.PurchaseOrder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateRequest(tenantDB, tenantID, userID, &req); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"message": "Purchase request created"})
}

func (h *PurchaseHandler) PurchaserAction(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	tenantID := c.MustGet("tenantID").(uint)
	userID := c.MustGet("userID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.PurchaserAction(tenantDB, tenantID, uint(id), userID, req.Action); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Order " + req.Action + "ed successfully"})
}

// 3. Receive Order (Stock Manager)
func (h *PurchaseHandler) Receive(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	tenantID := c.MustGet("tenantID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.ReceiveOrder(tenantDB, tenantID, uint(id)); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Order received and inventory updated"})
}

func (h *PurchaseHandler) UpdateRequest(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	tenantID := c.MustGet("tenantID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Quantity int     `json:"quantity"`
		BuyPrice float64 `json:"buy_price"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateRequest(tenantDB, tenantID, uint(id), req.Quantity, req.BuyPrice); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Request updated and resent"})
}

func (h *PurchaseHandler) List(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	tenantID := c.MustGet("tenantID").(uint)
	status := c.Query("status")
	orders, err := h.service.ListOrders(tenantDB, tenantID, status)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": orders})
}
