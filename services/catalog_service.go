package services

import (
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type CatalogService struct{}

func NewCatalogService() *CatalogService {
	return &CatalogService{}
}

// === Product Logic ===

func (s *CatalogService) CreateProduct(tenantDB *gorm.DB, tenantID uint, product *models.Product) error {
	repo := repositories.NewCatalogRepository(tenantDB)

	// 1. ✅ PLAN LIMIT CHECK (Products)
	var tenant models.Tenant
	config.GetMasterDB().Preload("Plan").First(&tenant, tenantID)

	if tenant.Plan.MaxProducts > 0 {
		currentCount, _ := repo.CountProducts()
		if int(currentCount) >= tenant.Plan.MaxProducts {
			// ✅ FIX: Removed '!' and simplified the message
			return fmt.Errorf("plan limit reached: you can only add %d products", tenant.Plan.MaxProducts)
		}
	}

	// 2. Create Product
	product.TenantID = tenantID
	return repo.CreateProduct(product)
}

func (s *CatalogService) ListProducts(tenantDB *gorm.DB, page, pageSize int) ([]models.Product, int64, error) {
	repo := repositories.NewCatalogRepository(tenantDB)
	offset := (page - 1) * pageSize
	return repo.ListProducts(offset, pageSize)
}

// === Category Logic ===

func (s *CatalogService) CreateCategory(tenantDB *gorm.DB, tenantID uint, category *models.Category) error {
	category.TenantID = tenantID
	return repositories.NewCatalogRepository(tenantDB).CreateCategory(category)
}

func (s *CatalogService) ListCategories(tenantDB *gorm.DB) ([]models.Category, error) {
	return repositories.NewCatalogRepository(tenantDB).ListCategories()
}
