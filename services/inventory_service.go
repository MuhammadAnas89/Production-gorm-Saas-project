package services

import (
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type InventoryService struct{}

func NewInventoryService() *InventoryService {
	return &InventoryService{}
}

// Stock update karna (e.g., jab order place ho)
func (s *InventoryService) UpdateStock(tenantDB *gorm.DB, productID uint, tenantID uint, quantity int) error {
	repo := repositories.NewInventoryRepository(tenantDB)
	return repo.UpdateStock(productID, tenantID, quantity)
}

func (s *InventoryService) GetLowStockAlerts(tenantDB *gorm.DB, tenantID uint, threshold int) ([]models.Inventory, error) {
	repo := repositories.NewInventoryRepository(tenantDB)
	return repo.GetLowStockProducts(tenantID, threshold)
}
