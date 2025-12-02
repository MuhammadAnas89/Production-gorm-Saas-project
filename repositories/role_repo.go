package repositories

import (
	"fmt"
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *models.Role) error
	List() ([]models.Role, error)
	GetByID(id uint) (*models.Role, error)
	AssignPermissions(roleID uint, permIDs []uint) error
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

func (r *roleRepository) List() ([]models.Role, error) {
	fmt.Println(">>> DEBUG: Main NAYA code hoon, Preload Permissions chala raha hoon <<<")
	var roles []models.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

func (r *roleRepository) AssignPermissions(roleID uint, permIDs []uint) error {
	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var perms []models.Permission
	if err := r.db.Where("id IN ?", permIDs).Find(&perms).Error; err != nil {
		return err
	}

	return r.db.Model(&role).Association("Permissions").Replace(&perms)
}

func (r *roleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}
func (r *roleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}
