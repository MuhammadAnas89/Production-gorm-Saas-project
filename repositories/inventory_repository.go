package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	UpdateStock(productID uint, tenantID uint, quantity int) error
	GetLowStockProducts(tenantID uint, threshold int) ([]models.Inventory, error)
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) UpdateStock(productID uint, tenantID uint, quantity int) error {

	return r.db.Model(&models.Inventory{}).
		Where("product_id = ? AND tenant_id = ?", productID, tenantID).
		Update("quantity", quantity).Error
}

func (r *inventoryRepository) GetLowStockProducts(tenantID uint, threshold int) ([]models.Inventory, error) {
	var inv []models.Inventory

	err := r.db.Preload("Product").
		Where("tenant_id = ? AND quantity <= ?", tenantID, threshold).
		Find(&inv).Error
	return inv, err
}
