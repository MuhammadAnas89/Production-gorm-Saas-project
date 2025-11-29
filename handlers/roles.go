package handlers

import (
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// 1. Create Role
func (h *RoleHandler) CreateRole(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	tenantID := c.MustGet("tenantID").(uint)

	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleService.CreateRole(tenantDB, tenantID, &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Role created", "role": role})
}

// 2. List Roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	roles, err := h.roleService.ListRoles(tenantDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// 3. Get Role by ID
func (h *RoleHandler) GetRole(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	id, _ := strconv.Atoi(c.Param("id"))

	role, err := h.roleService.GetRole(tenantDB, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": role})
}

// 4. Update Permissions for a Role
func (h *RoleHandler) UpdatePermissions(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	roleID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleService.UpdateRolePermissions(tenantDB, uint(roleID), req.PermissionIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions updated successfully"})
}
