package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type ModuleRepository interface {
	Create(module *models.Module) error
	List() ([]models.Module, error)
	GetByID(id uint) (*models.Module, error)
	Update(module *models.Module) error
	Delete(id uint) error
}

type moduleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) Create(module *models.Module) error {
	return r.db.Create(module).Error
}

func (r *moduleRepository) List() ([]models.Module, error) {
	var modules []models.Module
	err := r.db.Preload("Permissions").Find(&modules).Error
	return modules, err
}

func (r *moduleRepository) GetByID(id uint) (*models.Module, error) {
	var module models.Module
	err := r.db.Preload("Permissions").First(&module, id).Error
	return &module, err
}

func (r *moduleRepository) Update(module *models.Module) error {
	return r.db.Save(module).Error
}

func (r *moduleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Module{}, id).Error
}
