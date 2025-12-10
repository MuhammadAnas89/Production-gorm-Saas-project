package services

import (
	"errors"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}

func (s *RoleService) CreateRole(tenantDB *gorm.DB, tenantID uint, role *models.Role) error {

	role.TenantID = tenantID
	role.IsSystemRole = false

	repo := repositories.NewRoleRepository(tenantDB)
	return repo.Create(role)
}

func (s *RoleService) ListRoles(tenantDB *gorm.DB, tenantID uint) ([]models.Role, error) {
	repo := repositories.NewRoleRepository(tenantDB)
	return repo.List(tenantID)
}

func (s *RoleService) GetRole(tenantDB *gorm.DB, id uint) (*models.Role, error) {
	repo := repositories.NewRoleRepository(tenantDB)
	return repo.GetByID(id)
}

func (s *RoleService) UpdateRolePermissions(tenantDB *gorm.DB, roleID uint, permissionIDs []uint) error {
	repo := repositories.NewRoleRepository(tenantDB)

	role, err := repo.GetByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystemRole {

		return errors.New("cannot modify system roles")
	}

	return repo.AssignPermissions(roleID, permissionIDs)
}
