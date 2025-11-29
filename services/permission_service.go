package services

import (
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type PermissionService struct{}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

func (s *PermissionService) Create(masterDB *gorm.DB, perm *models.Permission) error {
	return repositories.NewPermissionRepository(masterDB).Create(perm)
}

func (s *PermissionService) List(masterDB *gorm.DB) ([]models.Permission, error) {
	return repositories.NewPermissionRepository(masterDB).List()
}

func (s *PermissionService) Delete(masterDB *gorm.DB, id uint) error {
	return repositories.NewPermissionRepository(masterDB).Delete(id)
}
func (s *PermissionService) UpdatePermiss(masterDB *gorm.DB, roleID uint, permIDs []uint) error {
	return repositories.NewRoleRepository(masterDB).AssignPermissions(roleID, permIDs)
}
