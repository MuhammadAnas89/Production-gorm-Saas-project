package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type ModuleRepository interface {
	Create(m *models.Module) error
	GetByID(id uint) (*models.Module, error)
	GetByName(name string) (*models.Module, error)
	List() ([]models.Module, error)
	Update(m *models.Module) error
	Delete(id uint) error
}

type moduleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) Create(m *models.Module) error {
	return r.db.Create(m).Error
}

func (r *moduleRepository) GetByID(id uint) (*models.Module, error) {
	var m models.Module
	err := r.db.First(&m, id).Error
	return &m, err
}

func (r *moduleRepository) GetByName(name string) (*models.Module, error) {
	var m models.Module
	err := r.db.Where("name = ?", name).First(&m).Error
	return &m, err
}

func (r *moduleRepository) List() ([]models.Module, error) {
	var mods []models.Module
	err := r.db.Find(&mods).Error
	return mods, err
}

func (r *moduleRepository) Update(m *models.Module) error {
	return r.db.Save(m).Error
}

func (r *moduleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Module{}, id).Error
}
