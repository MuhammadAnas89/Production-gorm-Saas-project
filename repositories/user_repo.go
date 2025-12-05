package repositories

import (
	"go-multi-tenant/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(offset, limit int) ([]models.User, int64, error)
	Count() (int64, error)

	// Role Management
	GetRoleByID(roleID uint) (*models.Role, error) // ✅ Added to keep Service clean
	AssignRole(userID uint, roleID uint) error
	RemoveRole(userID uint, roleID uint) error
	ReplaceRole(userID uint, roleID uint) error
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

	query := r.db.Preload("Roles")
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Find(&users).Error
	return users, count, err
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// ✅ New Helper to fetch Role inside Repo
func (r *userRepository) GetRoleByID(roleID uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, roleID).Error
	return &role, err
}

func (r *userRepository) AssignRole(userID uint, roleID uint) error {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}

	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	// Association check to avoid duplicates
	count := r.db.Model(&user).Association("Roles").Count()
	if count > 0 {
		// Optional: Clear existing roles if user can have only one role,
		// otherwise just append. Let's assume append for now.
	}

	return r.db.Model(&user).Association("Roles").Append(&role)
}

func (r *userRepository) RemoveRole(userID uint, roleID uint) error {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}
	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}
	return r.db.Model(&user).Association("Roles").Delete(&role)
}
func (r *userRepository) ReplaceRole(userID uint, roleID uint) error {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}

	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	// DB Logic yahan chipa di
	// Pehle clear karo, phir naya lagao
	if err := r.db.Model(&user).Association("Roles").Clear(); err != nil {
		return err
	}
	return r.db.Model(&user).Association("Roles").Append(&role)
}
