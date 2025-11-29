package handlers

import (
	"go-multi-tenant/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService *services.TenantService
}

func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req services.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Service Call (Creates Tenant, DB, Admin & Permissions)
	tenant, err := h.tenantService.CreateTenant(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Tenant created successfully",
		"tenant_id": tenant.ID,
		"db_name":   tenant.DBName,
	})
}

func (h *TenantHandler) ListTenants(c *gin.Context) {
	tenants, err := h.tenantService.ListTenants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tenants})
}
