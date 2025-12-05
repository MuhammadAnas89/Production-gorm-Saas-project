package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(tenant *models.Tenant) error
	GetByID(id uint) (*models.Tenant, error)
	GetGlobalIdentity(email string) (*models.GlobalIdentity, error)
	GetTenantWithPlan(id uint) (*models.Tenant, error)
	CreateGlobalIdentity(identity *models.GlobalIdentity) error
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

	err := r.db.Preload("Plan").First(&tenant, id).Error
	return &tenant, err
}
func (r *tenantRepository) GetTenantWithPlan(id uint) (*models.Tenant, error) {
	var tenant models.Tenant

	err := r.db.Preload("Plan").First(&tenant, id).Error
	return &tenant, err
}
func (r *tenantRepository) GetGlobalIdentity(email string) (*models.GlobalIdentity, error) {
	var identity models.GlobalIdentity
	err := r.db.Where("email = ?", email).First(&identity).Error
	return &identity, err
}

func (r *tenantRepository) CreateGlobalIdentity(identity *models.GlobalIdentity) error {
	return r.db.Create(identity).Error
}
