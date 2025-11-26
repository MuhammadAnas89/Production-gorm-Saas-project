package handlers

import (
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService *services.TenantService
}

func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

type CreateTenantRequest struct {
	Name          string              `json:"name" binding:"required"`
	DatabaseType  models.DatabaseType `json:"database_type" binding:"required"`
	AdminUsername string              `json:"admin_username" binding:"required"`
	AdminEmail    string              `json:"admin_email" binding:"required,email"`
	AdminPassword string              `json:"admin_password" binding:"required,min=6"`
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	response, err := h.tenantService.CreateTenant(&services.CreateTenantRequest{
		Name:          req.Name,
		DatabaseType:  req.DatabaseType,
		AdminUsername: req.AdminUsername,
		AdminEmail:    req.AdminEmail,
		AdminPassword: req.AdminPassword,
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create tenant", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Tenant created successfully", response)
}

func (h *TenantHandler) ListTenants(c *gin.Context) {
	tenants, err := h.tenantService.ListTenants()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list tenants", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tenants retrieved successfully", tenants)
}
