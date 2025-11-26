package services

import (
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"

	"gorm.io/gorm"
)

type ModuleService struct {
	// ❌ FIX: Repo removed
}

func NewModuleService() *ModuleService {
	return &ModuleService{}
}

// ✅ FIX: Methods now accept DB connection
func (s *ModuleService) Create(db *gorm.DB, m *models.Module) error {
	return repositories.NewModuleRepository(db).Create(m)
}

func (s *ModuleService) GetByID(db *gorm.DB, id uint) (*models.Module, error) {
	return repositories.NewModuleRepository(db).GetByID(id)
}

func (s *ModuleService) List(db *gorm.DB) ([]models.Module, error) {
	return repositories.NewModuleRepository(db).List()
}

func (s *ModuleService) Update(db *gorm.DB, m *models.Module) error {
	return repositories.NewModuleRepository(db).Update(m)
}

func (s *ModuleService) Delete(db *gorm.DB, id uint) error {
	return repositories.NewModuleRepository(db).Delete(id)
}
