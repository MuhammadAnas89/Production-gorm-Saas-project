package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	Create(perm *models.Permission) error
	List() ([]models.Permission, error)
	GetByCategory(category string) ([]models.Permission, error)
	Delete(id uint) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(perm *models.Permission) error {
	return r.db.Create(perm).Error
}

func (r *permissionRepository) List() ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.Find(&perms).Error
	return perms, err
}

func (r *permissionRepository) GetByCategory(category string) ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.Where("category = ?", category).Find(&perms).Error
	return perms, err
}

func (r *permissionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Permission{}, id).Error
}
