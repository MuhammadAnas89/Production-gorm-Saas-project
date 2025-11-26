package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(tenant *models.Tenant) error
	GetByID(id uint) (*models.Tenant, error)
	GetByName(name string) (*models.Tenant, error)
	List() ([]models.Tenant, error)
	Update(tenant *models.Tenant) error
	Delete(id uint) error
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(tenant *models.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *tenantRepository) GetByID(id uint) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.First(&tenant, id).Error
	return &tenant, err
}

func (r *tenantRepository) GetByName(name string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.Where("name = ?", name).First(&tenant).Error
	return &tenant, err
}

func (r *tenantRepository) List() ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := r.db.Find(&tenants).Error
	return tenants, err
}

func (r *tenantRepository) Update(tenant *models.Tenant) error {
	return r.db.Save(tenant).Error
}

func (r *tenantRepository) Delete(id uint) error {
	return r.db.Delete(&models.Tenant{}, id).Error
}
