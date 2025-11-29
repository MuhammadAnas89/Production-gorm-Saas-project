package handlers

import (
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ModuleHandler struct {
	modService *services.ModuleService
}

func NewModuleHandler(modService *services.ModuleService) *ModuleHandler {
	return &ModuleHandler{modService: modService}
}

// 1. Create Module
func (h *ModuleHandler) Create(c *gin.Context) {
	var module models.Module
	if err := c.ShouldBindJSON(&module); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Master DB pass kar rahe hain
	if err := h.modService.Create(config.MasterDB, &module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Module created", "data": module})
}

// 2. List Modules
func (h *ModuleHandler) List(c *gin.Context) {
	modules, err := h.modService.List(config.MasterDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": modules})
}

// 3. Update Module
func (h *ModuleHandler) Update(c *gin.Context) {
	var module models.Module
	if err := c.ShouldBindJSON(&module); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// ID from URL should match or be handled logic wise, assuming ID is in body or handled by repo update logic

	if err := h.modService.Update(config.MasterDB, &module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Module updated"})
}

// 4. Delete Module
func (h *ModuleHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.modService.Delete(config.MasterDB, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Module deleted"})
}
