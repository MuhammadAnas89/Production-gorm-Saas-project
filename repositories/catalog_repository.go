package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type CatalogRepository interface {
	CreateCategory(cat *models.Category) error
	ListCategories(tenantID uint) ([]models.Category, error)
	CreateProduct(product *models.Product) error
	GetProductByID(id uint, tenantID uint) (*models.Product, error)
	ListProducts(tenantID uint, offset, limit int) ([]models.Product, int64, error)
	CountProducts(tenantID uint) (int64, error)
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
	err := r.db.Where("tenant_id = ?", tenantID).Find(&cats).Error
	return cats, err
}

func (r *catalogRepository) CreateProduct(product *models.Product) error {

	return r.db.Create(product).Error
}

func (r *catalogRepository) GetProductByID(id uint, tenantID uint) (*models.Product, error) {
	var p models.Product
	err := r.db.Preload("Category").Preload("Inventory").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&p).Error
	return &p, err
}

func (r *catalogRepository) ListProducts(tenantID uint, offset, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var count int64
	r.db.Model(&models.Product{}).Where("tenant_id = ?", tenantID).Count(&count)

	err := r.db.Preload("Category").Preload("Inventory").
		Where("tenant_id = ?", tenantID).
		Offset(offset).Limit(limit).
		Find(&products).Error

	return products, count, err
}

func (r *catalogRepository) CountProducts(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
