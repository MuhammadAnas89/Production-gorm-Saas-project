package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type PurchaseRepository interface {
	Create(po *models.PurchaseOrder) error
	GetByID(id uint, tenantID uint) (*models.PurchaseOrder, error)
	List(tenantID uint, status string) ([]models.PurchaseOrder, error)
	Update(po *models.PurchaseOrder) error
}

type purchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) PurchaseRepository {
	return &purchaseRepository{db: db}
}

func (r *purchaseRepository) Create(po *models.PurchaseOrder) error {
	return r.db.Create(po).Error
}

func (r *purchaseRepository) GetByID(id uint, tenantID uint) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder

	err := r.db.Preload("Product").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&po).Error
	return &po, err
}

func (r *purchaseRepository) List(tenantID uint, status string) ([]models.PurchaseOrder, error) {
	var orders []models.PurchaseOrder
	query := r.db.Preload("Product").Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&orders).Error
	return orders, err
}

func (r *purchaseRepository) Update(po *models.PurchaseOrder) error {
	return r.db.Save(po).Error
}
