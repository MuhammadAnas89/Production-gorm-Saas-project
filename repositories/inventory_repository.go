package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	UpdateStock(productID uint, quantity int) error
	GetLowStockProducts(threshold int) ([]models.Inventory, error)
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) UpdateStock(productID uint, quantity int) error {

	return r.db.Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Update("quantity", quantity).Error
}

func (r *inventoryRepository) GetLowStockProducts(threshold int) ([]models.Inventory, error) {
	var inv []models.Inventory

	err := r.db.Preload("Product").Where("quantity <= ?", threshold).Find(&inv).Error
	return inv, err
}
