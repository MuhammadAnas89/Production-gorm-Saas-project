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

func (s *CatalogService) CreateProduct(tenantDB *gorm.DB, tenantID uint, product *models.Product) error {
	repo := repositories.NewCatalogRepository(tenantDB)

	var tenant models.Tenant
	config.GetMasterDB().Preload("Plan").First(&tenant, tenantID)

	if tenant.Plan.MaxProducts > 0 {
		currentCount, _ := repo.CountProducts(tenantID)
		if int(currentCount) >= tenant.Plan.MaxProducts {
			return fmt.Errorf("plan limit reached: you can only add %d products", tenant.Plan.MaxProducts)
		}
	}

	if product.Inventory == nil {
		product.Inventory = &models.Inventory{
			TenantID:      tenantID, // Important: TenantID link karna
			Quantity:      0,
			LowStockAlert: 10,
			Location:      "Main Warehouse",
		}
	} else {

		product.Inventory.TenantID = tenantID
	}

	product.TenantID = tenantID
	return repo.CreateProduct(product)
}
func (s *CatalogService) ListProducts(tenantDB *gorm.DB, tenantID uint, page, pageSize int) ([]models.Product, int64, error) {
	repo := repositories.NewCatalogRepository(tenantDB)
	offset := (page - 1) * pageSize

	return repo.ListProducts(tenantID, offset, pageSize)
}

func (s *CatalogService) CreateCategory(tenantDB *gorm.DB, tenantID uint, category *models.Category) error {
	category.TenantID = tenantID
	return repositories.NewCatalogRepository(tenantDB).CreateCategory(category)
}
func (s *CatalogService) ListCategories(tenantDB *gorm.DB, tenantID uint) ([]models.Category, error) {

	return repositories.NewCatalogRepository(tenantDB).ListCategories(tenantID)
}
