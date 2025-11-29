package handlers

import (
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CatalogHandler struct {
	catalogService *services.CatalogService
}

func NewCatalogHandler(catalogService *services.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

// === PRODUCTS ===

func (h *CatalogHandler) CreateProduct(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	tenantID := c.MustGet("tenantID").(uint)

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.catalogService.CreateProduct(tenantDB, tenantID, &product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created", "data": product})
}

func (h *CatalogHandler) ListProducts(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	products, count, err := h.catalogService.ListProducts(tenantDB, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      products,
		"total":     count,
		"page":      page,
		"page_size": pageSize,
	})
}

// === CATEGORIES ===

func (h *CatalogHandler) CreateCategory(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	tenantID := c.MustGet("tenantID").(uint)

	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.catalogService.CreateCategory(tenantDB, tenantID, &category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Category created", "data": category})
}

func (h *CatalogHandler) ListCategories(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	cats, err := h.catalogService.ListCategories(tenantDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cats})
}
