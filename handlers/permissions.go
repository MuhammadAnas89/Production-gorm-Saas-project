package handlers

import (
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permService *services.PermissionService
}

func NewPermissionHandler(permService *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService: permService}
}

func (h *PermissionHandler) Create(c *gin.Context) {
	var perm models.Permission
	if err := c.ShouldBindJSON(&perm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.permService.Create(config.MasterDB, &perm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Permission created", "data": perm})
}

func (h *PermissionHandler) List(c *gin.Context) {
	perms, err := h.permService.List(config.MasterDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.permService.Delete(config.MasterDB, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted"})
}
