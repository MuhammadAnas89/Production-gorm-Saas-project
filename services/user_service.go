package services

import (
	"errors"
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"go-multi-tenant/utils"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   uint   `json:"role_id"`
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isStrongPassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "password must be at least 8 characters long"
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return false, "password must contain uppercase, lowercase, number, and special character"
	}
	return true, ""
}

func clearUserCache(tenantID uint) {
	cacheKey := fmt.Sprintf("tenant:%d:users:list", tenantID)
	_ = config.DeleteCache(cacheKey)
}

func (s *UserService) CreateUser(tenantDB *gorm.DB, tenantID uint, req *CreateUserRequest, currentUser *models.User) (*models.User, error) {
	if !currentUser.HasPermission("user:create") {
		return nil, errors.New("insufficient permissions")
	}
	if !isValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}
	if strong, msg := isStrongPassword(req.Password); !strong {
		return nil, errors.New(msg)
	}
	if len(req.Username) < 3 {
		return nil, errors.New("username must be at least 3 characters long")
	}

	userRepo := repositories.NewUserRepository(tenantDB)

	existingUser, err := userRepo.GetByUsernameAndTenant(req.Username, tenantID)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists in this tenant")
	}
	existingEmailUser, err := userRepo.GetByEmail(req.Email, tenantID)
	if err == nil && existingEmailUser != nil {
		return nil, errors.New("email already exists in this tenant")
	}

	var role models.Role
	if err := tenantDB.Preload("Permissions").First(&role, req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to fetch role: %w", err)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		TenantID: tenantID,
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		APIKey:   "",
		IsActive: true,
	}

	err = userRepo.Create(user)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errors.New("user already exists (duplicate entry)")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tenantDB.Model(user).Association("Roles").Append(&role); err != nil {
		return nil, fmt.Errorf("failed to assign role to user: %w", err)
	}

	clearUserCache(tenantID)

	var userWithRoles models.User
	if err := tenantDB.Preload("Roles").First(&userWithRoles, user.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user with roles: %w", err)
	}

	return &userWithRoles, nil
}

func (s *UserService) GetUser(tenantDB *gorm.DB, userID uint, currentUser *models.User) (*models.User, error) {
	userRepo := repositories.NewUserRepository(tenantDB)

	if currentUser.HasRole("Tenant User") {
		if currentUser.ID != userID {
			return nil, errors.New("access denied - can only view own profile")
		}
	}

	user, err := userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if user.TenantID != currentUser.TenantID {
		return nil, errors.New("access denied")
	}

	var userWithRoles models.User
	if err := tenantDB.Preload("Roles").First(&userWithRoles, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user roles: %w", err)
	}

	return &userWithRoles, nil
}

func (s *UserService) UpdateUser(tenantDB *gorm.DB, userID uint, updateData map[string]interface{}, currentUser *models.User) (*models.User, error) {
	userRepo := repositories.NewUserRepository(tenantDB)

	if !currentUser.HasPermission("user:update") {
		return nil, errors.New("insufficient permissions")
	}

	user, err := userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if user.TenantID != currentUser.TenantID {
		return nil, errors.New("access denied")
	}

	if username, exists := updateData["username"]; exists {
		existingUser, err := userRepo.GetByUsernameAndTenant(username.(string), currentUser.TenantID)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("username already exists in this tenant")
		}
		user.Username = username.(string)
	}
	if email, exists := updateData["email"]; exists {
		emailStr := email.(string)
		if !isValidEmail(emailStr) {
			return nil, errors.New("invalid email format")
		}
		existingEmailUser, err := userRepo.GetByEmail(emailStr, currentUser.TenantID)
		if err == nil && existingEmailUser != nil && existingEmailUser.ID != userID {
			return nil, errors.New("email already exists in this tenant")
		}
		user.Email = emailStr
	}
	if isActive, exists := updateData["is_active"]; exists {
		user.IsActive = isActive.(bool)
	}

	if roleID, exists := updateData["role_id"]; exists {
		var role models.Role
		var rID uint
		switch v := roleID.(type) {
		case float64:
			rID = uint(v)
		case uint:
			rID = v
		case int:
			rID = uint(v)
		}

		if err := tenantDB.First(&role, rID).Error; err == nil {
			tenantDB.Model(user).Association("Roles").Clear()
			tenantDB.Model(user).Association("Roles").Append(&role)
		}
	}

	if err := userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	clearUserCache(currentUser.TenantID)

	return s.GetUser(tenantDB, userID, currentUser)
}

func (s *UserService) DeleteUser(tenantDB *gorm.DB, userID uint, currentUser *models.User) error {
	userRepo := repositories.NewUserRepository(tenantDB)

	if !currentUser.HasPermission("user:delete") {
		return errors.New("insufficient permissions")
	}

	if currentUser.ID == userID {
		return errors.New("cannot delete your own account")
	}

	user, err := userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if user.TenantID != currentUser.TenantID {
		return errors.New("access denied")
	}

	if err := userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	clearUserCache(currentUser.TenantID)

	return nil
}

func (s *UserService) ListUsers(tenantDB *gorm.DB, currentUser *models.User) ([]models.User, error) {
	if !currentUser.HasPermission("user:list") && !currentUser.HasPermission("user:read") {
		return nil, errors.New("insufficient permissions")
	}

	cacheKey := fmt.Sprintf("tenant:%d:users:list", currentUser.TenantID)
	var cachedUsers []models.User

	err := config.GetCacheStruct(cacheKey, &cachedUsers)
	if err == nil {
		fmt.Println("🔥 REDIS CACHE HIT! Serving from Memory")
		return cachedUsers, nil
	}

	fmt.Println("🐢 REDIS MISS! Fetching from DB...")

	userRepo := repositories.NewUserRepository(tenantDB)
	_, err = userRepo.ListByTenant(currentUser.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	var usersWithRoles []models.User
	if err := tenantDB.Preload("Roles").Where("tenant_id = ?", currentUser.TenantID).Find(&usersWithRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to load users with roles: %w", err)
	}

	var userResponses []models.User
	for _, user := range usersWithRoles {
		userResponses = append(userResponses, models.User{
			ID:        user.ID,
			TenantID:  user.TenantID,
			Username:  user.Username,
			Email:     user.Email,
			Roles:     user.Roles,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
		})
	}

	_ = config.SetCacheStruct(cacheKey, userResponses, 0)

	return userResponses, nil
}

func (s *UserService) GetSelf(tenantDB *gorm.DB, currentUser *models.User) (*models.User, error) {
	userRepo := repositories.NewUserRepository(tenantDB)

	_, err := userRepo.GetByID(currentUser.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	var userWithRoles models.User
	if err := tenantDB.Preload("Roles.Permissions").First(&userWithRoles, currentUser.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user data: %w", err)
	}

	return &userWithRoles, nil
}

func (s *UserService) GetAvailableRoles(tenantDB *gorm.DB, currentUser *models.User) ([]models.Role, error) {
	if !currentUser.HasPermission("user:create") {
		return nil, errors.New("insufficient permissions")
	}

	var roles []models.Role
	if err := tenantDB.Preload("Permissions").Where("tenant_id = ?", currentUser.TenantID).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	return roles, nil
}

func (s *UserService) AssignRole(tenantDB *gorm.DB, userID uint, roleID uint, currentUser *models.User) error {
	if !currentUser.HasPermission("user:update") {
		return errors.New("insufficient permissions")
	}

	userRepo := repositories.NewUserRepository(tenantDB)
	user, err := userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if user.TenantID != currentUser.TenantID {
		return errors.New("access denied")
	}

	var role models.Role
	if err := tenantDB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	var count int64
	tenantDB.Table("user_roles").Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count)
	if count > 0 {
		return errors.New("role already assigned")
	}

	if err := tenantDB.Model(user).Association("Roles").Append(&role); err != nil {
		return err
	}

	clearUserCache(currentUser.TenantID)

	return nil
}

func (s *UserService) RemoveRole(tenantDB *gorm.DB, userID uint, roleID uint, currentUser *models.User) error {
	if !currentUser.HasPermission("user:update") {
		return errors.New("insufficient permissions")
	}

	userRepo := repositories.NewUserRepository(tenantDB)
	user, err := userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if user.TenantID != currentUser.TenantID {
		return errors.New("access denied")
	}

	var role models.Role
	if err := tenantDB.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found")
	}

	if err := tenantDB.Model(user).Association("Roles").Delete(&role); err != nil {
		return err
	}

	clearUserCache(currentUser.TenantID)

	return nil
}
