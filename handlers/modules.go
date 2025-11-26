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

type ModuleHandler struct {
	svc *services.ModuleService
}

func NewModuleHandler(svc *services.ModuleService) *ModuleHandler {
	return &ModuleHandler{svc: svc}
}

func (h *ModuleHandler) CreateModule(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	var m models.Module
	if err := c.ShouldBindJSON(&m); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid body", err)
		return
	}
	if err := h.svc.Create(tenantDB, &m); err != nil { // ✅ Pass DB
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to create module", err)
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "module created", m)
}

func (h *ModuleHandler) GetModule(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	m, err := h.svc.GetByID(tenantDB, uint(id64)) // ✅ Pass DB
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "module not found", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "module retrieved", m)
}

func (h *ModuleHandler) ListModules(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	mods, err := h.svc.List(tenantDB)            // ✅ Pass DB
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to list modules", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "modules list", mods)
}

func (h *ModuleHandler) UpdateModule(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	var m models.Module
	if err := c.ShouldBindJSON(&m); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid body", err)
		return
	}
	m.ID = uint(id64)
	if err := h.svc.Update(tenantDB, &m); err != nil { // ✅ Pass DB
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to update module", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "module updated", m)
}

func (h *ModuleHandler) DeleteModule(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB) // ✅ Get DB
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	if err := h.svc.Delete(tenantDB, uint(id64)); err != nil { // ✅ Pass DB
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to delete module", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "module deleted", nil)
}
