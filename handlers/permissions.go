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

type PermissionHandler struct {
	svc *services.PermissionService
}

func NewPermissionHandler(svc *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	var p models.Permission
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid body", err)
		return
	}
	if err := h.svc.Create(tenantDB, &p); err != nil { // ✅ Pass DB
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to create permission", err)
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "permission created", p)
}

func (h *PermissionHandler) GetPermission(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	p, err := h.svc.GetByID(tenantDB, uint(id64)) // ✅ Pass DB
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "permission not found", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "permission retrieved", p)
}

func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	perms, err := h.svc.List(tenantDB)           // ✅ Pass DB
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to list permissions", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "permissions list", perms)
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	var p models.Permission
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid body", err)
		return
	}
	p.ID = uint(id64)
	if err := h.svc.Update(tenantDB, &p); err != nil { // ✅ Pass DB
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to update permission", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "permission updated", p)
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	if err := h.svc.Delete(tenantDB, uint(id64)); err != nil { // ✅ Pass DB
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to delete permission", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "permission deleted", nil)
}
