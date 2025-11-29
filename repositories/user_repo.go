package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error) // ✅ Added
	Update(user *models.User) error
	Delete(id uint) error
	List(offset, limit int) ([]models.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles.Permissions").First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles").Where("email = ?", email).First(&user).Error
	return &user, err
}

// ✅ Added this method
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) List(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var count int64

	r.db.Model(&models.User{}).Count(&count)

	// Agar limit -1 hai to sab le aao
	query := r.db.Preload("Roles")
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Find(&users).Error
	return users, count, err
}
