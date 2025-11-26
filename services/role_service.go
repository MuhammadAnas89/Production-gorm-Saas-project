package services

import (
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type RoleService struct {
	// ❌ FIX: Repo removed from struct to make it stateless
}

func NewRoleService() *RoleService {
	return &RoleService{}
}

// ✅ FIX: Methods now accept DB connection
func (s *RoleService) Create(db *gorm.DB, role *models.Role) error {
	return repositories.NewRoleRepository(db).Create(role)
}

func (s *RoleService) GetByID(db *gorm.DB, id uint) (*models.Role, error) {
	return repositories.NewRoleRepository(db).GetByID(id)
}

func (s *RoleService) List(db *gorm.DB) ([]models.Role, error) {
	return repositories.NewRoleRepository(db).List()
}

func (s *RoleService) Update(db *gorm.DB, role *models.Role) error {
	return repositories.NewRoleRepository(db).Update(role)
}

func (s *RoleService) Delete(db *gorm.DB, id uint) error {
	return repositories.NewRoleRepository(db).Delete(id)
}

func (s *RoleService) AddPermissions(db *gorm.DB, roleID uint, permissionIDs []uint) error {
	return repositories.NewRoleRepository(db).AddPermissions(roleID, permissionIDs)
}

func (s *RoleService) RemovePermission(db *gorm.DB, roleID uint, permissionID uint) error {
	return repositories.NewRoleRepository(db).RemovePermission(roleID, permissionID)
}
