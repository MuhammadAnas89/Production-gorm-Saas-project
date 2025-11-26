package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsernameAndTenant(username string, tenantID uint) (*models.User, error)
	GetByEmail(email string, tenantID uint) (*models.User, error)
	GetByAPIKey(apiKey string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	ListByTenant(tenantID uint) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {

	user.APIKey = ""
	user.APIKeyExpiry = nil
	return r.db.Create(user).Error
}

func (r *userRepository) GetByAPIKey(apiKey string) (*models.User, error) {
	var user models.User
	err := r.db.Where("api_key = ? AND api_key != ''", apiKey).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByUsernameAndTenant(username string, tenantID uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ? AND tenant_id = ?", username, tenantID).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByEmail(email string, tenantID uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	return &user, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) ListByTenant(tenantID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}
