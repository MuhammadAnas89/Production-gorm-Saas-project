package handlers

import (
	"net/http"
	"strconv"

	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	svc *services.RoleService
}

func NewRoleHandler(svc *services.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	IsSystem      bool   `json:"is_system_role"`
	PermissionIDs []uint `json:"permission_ids,omitempty"`
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid body", err)
		return
	}

	role := &models.Role{
		Name:         req.Name,
		Description:  req.Description,
		IsSystemRole: req.IsSystem,

		TenantID: currentUser.TenantID,
	}

	if err := h.svc.Create(tenantDB, role); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to create role", err)
		return
	}

	if len(req.PermissionIDs) > 0 {
		if err := h.svc.AddPermissions(tenantDB, role.ID, req.PermissionIDs); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "failed to attach permissions", err)
			return
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, "role created", role)
}

func (h *RoleHandler) GetRole(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)

	role, err := h.svc.GetByID(tenantDB, uint(id64))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "role not found", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "role retrieved", role)
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	roles, err := h.svc.List(tenantDB)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to list roles", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "roles list", roles)
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	var r models.Role
	if err := c.ShouldBindJSON(&r); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid body", err)
		return
	}
	r.ID = uint(id64)
	if err := h.svc.Update(tenantDB, &r); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to update role", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "role updated", r)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	if err := h.svc.Delete(tenantDB, uint(id64)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to delete role", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "role deleted", nil)
}

// Attach permissions to role
type AttachPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

func (h *RoleHandler) AttachPermissions(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	var req AttachPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid body", err)
		return
	}
	if err := h.svc.AddPermissions(tenantDB, uint(id64), req.PermissionIDs); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to attach permissions", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "permissions attached", nil)
}

func (h *RoleHandler) DetachPermission(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	pidStr := c.Param("pid")
	pid64, _ := strconv.ParseUint(pidStr, 10, 32)
	if err := h.svc.RemovePermission(tenantDB, uint(id64), uint(pid64)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to remove permission", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "permission removed", nil)
}
