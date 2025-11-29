package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

// Combined Repository for Product & Category
type CatalogRepository interface {
	// Category
	CreateCategory(cat *models.Category) error
	ListCategories() ([]models.Category, error)

	// Product
	CreateProduct(product *models.Product) error
	GetProductByID(id uint) (*models.Product, error)
	ListProducts(offset, limit int) ([]models.Product, int64, error)
	CountProducts() (int64, error) // For Plan Limit Check
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

func (r *catalogRepository) ListCategories() ([]models.Category, error) {
	var cats []models.Category
	err := r.db.Find(&cats).Error
	return cats, err
}

func (r *catalogRepository) CreateProduct(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *catalogRepository) GetProductByID(id uint) (*models.Product, error) {
	var p models.Product
	err := r.db.Preload("Category").Preload("Inventory").First(&p, id).Error
	return &p, err
}

func (r *catalogRepository) ListProducts(offset, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var count int64

	r.db.Model(&models.Product{}).Count(&count)
	err := r.db.Preload("Category").Preload("Inventory").Offset(offset).Limit(limit).Find(&products).Error
	return products, count, err
}

func (r *catalogRepository) CountProducts() (int64, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Count(&count).Error
	return count, err
}
