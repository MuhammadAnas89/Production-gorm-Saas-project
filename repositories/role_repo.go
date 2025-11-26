package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *models.Role) error
	GetByID(id uint) (*models.Role, error)
	GetByName(name string) (*models.Role, error)
	List() ([]models.Role, error)
	Update(role *models.Role) error
	Delete(id uint) error
	AddPermissions(roleID uint, permissionIDs []uint) error
	RemovePermission(roleID uint, permissionID uint) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Preload("Users").First(&role, id).Error
	return &role, err
}

func (r *roleRepository) GetByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *roleRepository) List() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}

func (r *roleRepository) AddPermissions(roleID uint, permissionIDs []uint) error {
	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var permissions []models.Permission
	if err := r.db.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return err
	}

	return r.db.Model(&role).Association("Permissions").Append(permissions)
}

func (r *roleRepository) RemovePermission(roleID uint, permissionID uint) error {
	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var permission models.Permission
	if err := r.db.First(&permission, permissionID).Error; err != nil {
		return err
	}

	return r.db.Model(&role).Association("Permissions").Delete(&permission)
}
