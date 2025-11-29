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
	// Custom role creation logic
	role.TenantID = tenantID
	role.IsSystemRole = false // Custom roles system roles nahi hote, delete ho sakte hain

	repo := repositories.NewRoleRepository(tenantDB)
	return repo.Create(role)
}

func (s *RoleService) ListRoles(tenantDB *gorm.DB) ([]models.Role, error) {
	repo := repositories.NewRoleRepository(tenantDB)
	return repo.List()
}

func (s *RoleService) GetRole(tenantDB *gorm.DB, id uint) (*models.Role, error) {
	repo := repositories.NewRoleRepository(tenantDB)
	return repo.GetByID(id)
}

// Permissions assign karna role ko
func (s *RoleService) UpdateRolePermissions(tenantDB *gorm.DB, roleID uint, permissionIDs []uint) error {
	repo := repositories.NewRoleRepository(tenantDB)

	// Check kar sakte hain ke role exist karta hai ya nahi
	role, err := repo.GetByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystemRole {
		// Optional: System roles (like Tenant Admin) ko edit karne se rokna hai ya nahi?
		// Usually System Roles fixed hote hain.
		return errors.New("cannot modify system roles")
	}

	return repo.AssignPermissions(roleID, permissionIDs)
}
