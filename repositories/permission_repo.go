package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	Create(p *models.Permission) error
	GetByID(id uint) (*models.Permission, error)
	GetByName(name string) (*models.Permission, error)
	List() ([]models.Permission, error)
	Update(p *models.Permission) error
	Delete(id uint) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(p *models.Permission) error {
	return r.db.Create(p).Error
}

func (r *permissionRepository) GetByID(id uint) (*models.Permission, error) {
	var p models.Permission
	err := r.db.First(&p, id).Error
	return &p, err
}

func (r *permissionRepository) GetByName(name string) (*models.Permission, error) {
	var p models.Permission
	err := r.db.Where("name = ?", name).First(&p).Error
	return &p, err
}

func (r *permissionRepository) List() ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.Find(&perms).Error
	return perms, err
}

func (r *permissionRepository) Update(p *models.Permission) error {
	return r.db.Save(p).Error
}

func (r *permissionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Permission{}, id).Error
}
