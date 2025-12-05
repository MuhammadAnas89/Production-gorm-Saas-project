package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type CatalogRepository interface {
	// Category
	CreateCategory(cat *models.Category) error
	ListCategories(tenantID uint) ([]models.Category, error) // ✅ Added TenantID

	// Product
	CreateProduct(product *models.Product) error
	GetProductByID(id uint, tenantID uint) (*models.Product, error)                 // ✅ Added TenantID check
	ListProducts(tenantID uint, offset, limit int) ([]models.Product, int64, error) // ✅ Added TenantID
	CountProducts(tenantID uint) (int64, error)                                     // ✅ Added TenantID
}

type catalogRepository struct {
	db *gorm.DB
}

func NewCatalogRepository(db *gorm.DB) CatalogRepository {
	return &catalogRepository{db: db}
}

func (r *catalogRepository) CreateCategory(cat *models.Category) error {
	return r.db.Create(cat).Error
}

func (r *catalogRepository) ListCategories(tenantID uint) ([]models.Category, error) {
	var cats []models.Category
	// ✅ Fix: Filter by TenantID
	err := r.db.Where("tenant_id = ?", tenantID).Find(&cats).Error
	return cats, err
}

func (r *catalogRepository) CreateProduct(product *models.Product) error {
	// Yahan se logic hata diya. Repo bas save karega.
	return r.db.Create(product).Error
}

func (r *catalogRepository) GetProductByID(id uint, tenantID uint) (*models.Product, error) {
	var p models.Product
	// ✅ Fix: Ensure Product belongs to Tenant
	err := r.db.Preload("Category").Preload("Inventory").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&p).Error
	return &p, err
}

func (r *catalogRepository) ListProducts(tenantID uint, offset, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var count int64

	// ✅ Fix: Count only this tenant's products
	r.db.Model(&models.Product{}).Where("tenant_id = ?", tenantID).Count(&count)

	err := r.db.Preload("Category").Preload("Inventory").
		Where("tenant_id = ?", tenantID). // ✅ Vital for Shared DB
		Offset(offset).Limit(limit).
		Find(&products).Error

	return products, count, err
}

func (r *catalogRepository) CountProducts(tenantID uint) (int64, error) {
	var count int64
	// ✅ Fix: Count only this tenant's products
	err := r.db.Model(&models.Product{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
