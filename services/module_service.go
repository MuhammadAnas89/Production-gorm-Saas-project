package services

import (
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type ModuleService struct{}

func NewModuleService() *ModuleService {
	return &ModuleService{}
}

func (s *ModuleService) Create(masterDB *gorm.DB, module *models.Module) error {
	return repositories.NewModuleRepository(masterDB).Create(module)
}

func (s *ModuleService) List(masterDB *gorm.DB) ([]models.Module, error) {
	return repositories.NewModuleRepository(masterDB).List()
}

func (s *ModuleService) Update(masterDB *gorm.DB, module *models.Module) error {
	return repositories.NewModuleRepository(masterDB).Update(module)
}

func (s *ModuleService) Delete(masterDB *gorm.DB, id uint) error {
	return repositories.NewModuleRepository(masterDB).Delete(id)
}
