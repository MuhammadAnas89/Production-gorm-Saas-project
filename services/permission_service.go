package services

import (
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type PermissionService struct {
}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

func (s *PermissionService) Create(db *gorm.DB, p *models.Permission) error {
	return repositories.NewPermissionRepository(db).Create(p)
}

func (s *PermissionService) GetByID(db *gorm.DB, id uint) (*models.Permission, error) {
	return repositories.NewPermissionRepository(db).GetByID(id)
}

func (s *PermissionService) List(db *gorm.DB) ([]models.Permission, error) {
	return repositories.NewPermissionRepository(db).List()
}

func (s *PermissionService) Update(db *gorm.DB, p *models.Permission) error {
	return repositories.NewPermissionRepository(db).Update(p)
}

func (s *PermissionService) Delete(db *gorm.DB, id uint) error {
	return repositories.NewPermissionRepository(db).Delete(id)
}
