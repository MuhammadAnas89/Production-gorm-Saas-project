package services

import (
	"errors"
	"fmt"
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

// ✅ Email validation helper
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ✅ Password strength validation helper
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

func (s *UserService) CreateUser(tenantDB *gorm.DB, tenantID uint, req *CreateUserRequest, currentUser *models.User) (*models.User, error) {
	// ✅ Permission check
	if !currentUser.HasPermission("user:create") {
		return nil, errors.New("insufficient permissions")
	}

	// ✅ FIX: Email validation
	if !isValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// ✅ FIX: Password strength validation
	if strong, msg := isStrongPassword(req.Password); !strong {
		return nil, errors.New(msg)
	}

	// ✅ Username validation
	if len(req.Username) < 3 {
		return nil, errors.New("username must be at least 3 characters long")
	}

	userRepo := repositories.NewUserRepository(tenantDB)

	// ✅ Check for existing username
	existingUser, err := userRepo.GetByUsernameAndTenant(req.Username, tenantID)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists in this tenant")
	}

	// ✅ Check for existing email
	existingEmailUser, err := userRepo.GetByEmail(req.Email, tenantID)
	if err == nil && existingEmailUser != nil {
		return nil, errors.New("email already exists in this tenant")
	}

	// ✅ FIX: Validate role exists and preload permissions
	var role models.Role
	if err := tenantDB.Preload("Permissions").First(&role, req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to fetch role: %w", err)
	}

	// ✅ Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// ✅ Create user
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
		// ✅ Handle duplicate entry errors
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "idx_username_tenant") {
				return nil, errors.New("username already exists in this tenant (database constraint)")
			}
			if strings.Contains(err.Error(), "idx_email_tenant") {
				return nil, errors.New("email already exists in this tenant (database constraint)")
			}
			if strings.Contains(err.Error(), "users.api_key") {
				user.APIKey = ""
				err = userRepo.Create(user)
				if err != nil {
					return nil, fmt.Errorf("failed to create user: %w", err)
				}
			} else {
				return nil, fmt.Errorf("duplicate entry error: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// ✅ Assign role to user
	if err := tenantDB.Model(user).Association("Roles").Append(&role); err != nil {
		return nil, fmt.Errorf("failed to assign role to user: %w", err)
	}

	// ✅ Load user with roles
	var userWithRoles models.User
	if err := tenantDB.Preload("Roles").First(&userWithRoles, user.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user with roles: %w", err)
	}

	// ✅ Return user response (without password)
	userResponse := &models.User{
		ID:        userWithRoles.ID,
		TenantID:  userWithRoles.TenantID,
		Username:  userWithRoles.Username,
		Email:     userWithRoles.Email,
		Roles:     userWithRoles.Roles,
		IsActive:  userWithRoles.IsActive,
		CreatedAt: userWithRoles.CreatedAt,
	}

	return userResponse, nil
}

func (s *UserService) GetUser(tenantDB *gorm.DB, userID uint, currentUser *models.User) (*models.User, error) {
	userRepo := repositories.NewUserRepository(tenantDB)

	// ✅ Tenant User can only view their own profile
	if !currentUser.HasRole("Tenant Admin") && currentUser.ID != userID {
		return nil, errors.New("you can only view your own profile")
	}
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

	// ✅ Ensure user belongs to same tenant
	if user.TenantID != currentUser.TenantID {
		return nil, errors.New("access denied")
	}

	// ✅ Load user with roles
	var userWithRoles models.User
	if err := tenantDB.Preload("Roles").First(&userWithRoles, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user roles: %w", err)
	}

	// ✅ Return user response (without password)
	userResponse := &models.User{
		ID:        userWithRoles.ID,
		TenantID:  userWithRoles.TenantID,
		Username:  userWithRoles.Username,
		Email:     userWithRoles.Email,
		Roles:     userWithRoles.Roles,
		IsActive:  userWithRoles.IsActive,
		CreatedAt: userWithRoles.CreatedAt,
	}

	return userResponse, nil
}

func (s *UserService) UpdateUser(tenantDB *gorm.DB, userID uint, updateData map[string]interface{}, currentUser *models.User) (*models.User, error) {
	userRepo := repositories.NewUserRepository(tenantDB)

	// ✅ Permission check
	if !currentUser.HasPermission("user:update") {
		return nil, errors.New("insufficient permissions")
	}

	user, err := userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// ✅ Ensure user belongs to same tenant
	if user.TenantID != currentUser.TenantID {
		return nil, errors.New("access denied")
	}

	// ✅ Validate username uniqueness
	if username, exists := updateData["username"]; exists {
		existingUser, err := userRepo.GetByUsernameAndTenant(username.(string), currentUser.TenantID)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("username already exists in this tenant")
		}
	}

	// ✅ Validate email uniqueness and format
	if email, exists := updateData["email"]; exists {
		emailStr := email.(string)
		if !isValidEmail(emailStr) {
			return nil, errors.New("invalid email format")
		}
		existingEmailUser, err := userRepo.GetByEmail(emailStr, currentUser.TenantID)
		if err == nil && existingEmailUser != nil && existingEmailUser.ID != userID {
			return nil, errors.New("email already exists in this tenant")
		}
	}

	// ✅ Update user fields
	if username, exists := updateData["username"]; exists {
		user.Username = username.(string)
	}
	if email, exists := updateData["email"]; exists {
		user.Email = email.(string)
	}
	if isActive, exists := updateData["is_active"]; exists {
		user.IsActive = isActive.(bool)
	}

	// ✅ Update role if provided
	if roleID, exists := updateData["role_id"]; exists {
		var role models.Role
		if err := tenantDB.Preload("Permissions").First(&role, roleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("role not found")
			}
			return nil, fmt.Errorf("failed to fetch role: %w", err)
		}

		if err := tenantDB.Model(user).Association("Roles").Clear(); err != nil {
			return nil, fmt.Errorf("failed to clear existing roles: %w", err)
		}
		if err := tenantDB.Model(user).Association("Roles").Append(&role); err != nil {
			return nil, fmt.Errorf("failed to assign new role: %w", err)
		}
	}

	// ✅ Save user updates
	if err := userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUser(tenantDB, userID, currentUser)
}

func (s *UserService) DeleteUser(tenantDB *gorm.DB, userID uint, currentUser *models.User) error {
	userRepo := repositories.NewUserRepository(tenantDB)

	// ✅ Permission check
	if !currentUser.HasPermission("user:delete") {
		return errors.New("insufficient permissions")
	}

	// ✅ Prevent self-deletion
	if currentUser.ID == userID {
		return errors.New("cannot delete your own account")
	}

	user, err := userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	// ✅ Ensure user belongs to same tenant
	if user.TenantID != currentUser.TenantID {
		return errors.New("access denied")
	}

	if err := userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *UserService) ListUsers(tenantDB *gorm.DB, currentUser *models.User) ([]models.User, error) {
	userRepo := repositories.NewUserRepository(tenantDB)

	// ✅ Permission check
	if !currentUser.HasPermission("user:read") {
		return nil, errors.New("insufficient permissions")
	}

	_, err := userRepo.ListByTenant(currentUser.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ✅ Load users with roles
	var usersWithRoles []models.User
	if err := tenantDB.Preload("Roles").Where("tenant_id = ?", currentUser.TenantID).Find(&usersWithRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to load users with roles: %w", err)
	}

	// ✅ Return user responses (without passwords)
	var userResponses []models.User
	for _, user := range usersWithRoles {
		userResponse := models.User{
			ID:        user.ID,
			TenantID:  user.TenantID,
			Username:  user.Username,
			Email:     user.Email,
			Roles:     user.Roles,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
		}
		userResponses = append(userResponses, userResponse)
	}

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

	// ✅ Load user with roles and permissions
	var userWithRoles models.User
	if err := tenantDB.Preload("Roles.Permissions").First(&userWithRoles, currentUser.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user data: %w", err)
	}

	// ✅ Return user response (without password)
	userResponse := &models.User{
		ID:        userWithRoles.ID,
		TenantID:  userWithRoles.TenantID,
		Username:  userWithRoles.Username,
		Email:     userWithRoles.Email,
		Roles:     userWithRoles.Roles,
		IsActive:  userWithRoles.IsActive,
		CreatedAt: userWithRoles.CreatedAt,
	}

	return userResponse, nil
}

func (s *UserService) GetAvailableRoles(tenantDB *gorm.DB, currentUser *models.User) ([]models.Role, error) {
	// ✅ Permission check
	if !currentUser.HasPermission("user:create") {
		return nil, errors.New("insufficient permissions")
	}

	var roles []models.Role
	if err := tenantDB.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	return roles, nil
}

func (s *UserService) AssignRole(tenantDB *gorm.DB, userID uint, roleID uint, currentUser *models.User) error {
	// ✅ Permission check
	if !currentUser.HasPermission("user:update") {
		return errors.New("insufficient permissions")
	}

	userRepo := repositories.NewUserRepository(tenantDB)
	user, err := userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	// ✅ Ensure user belongs to same tenant
	if user.TenantID != currentUser.TenantID {
		return errors.New("access denied - user from different tenant")
	}

	// ✅ Verify role exists
	var role models.Role
	if err := tenantDB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to fetch role: %w", err)
	}

	// ✅ Check if role is already assigned
	var existingRoles []models.Role
	if err := tenantDB.Model(user).Association("Roles").Find(&existingRoles); err != nil {
		return fmt.Errorf("failed to check existing roles: %w", err)
	}

	for _, r := range existingRoles {
		if r.ID == roleID {
			return errors.New("role already assigned to user")
		}
	}

	// ✅ Assign role
	if err := tenantDB.Model(user).Association("Roles").Append(&role); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

func (s *UserService) RemoveRole(tenantDB *gorm.DB, userID uint, roleID uint, currentUser *models.User) error {
	// ✅ Permission check
	if !currentUser.HasPermission("user:update") {
		return errors.New("insufficient permissions")
	}

	userRepo := repositories.NewUserRepository(tenantDB)
	user, err := userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	// ✅ Ensure user belongs to same tenant
	if user.TenantID != currentUser.TenantID {
		return errors.New("access denied - user from different tenant")
	}

	// ✅ Verify role exists
	var role models.Role
	if err := tenantDB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to fetch role: %w", err)
	}

	// ✅ Remove role
	if err := tenantDB.Model(user).Association("Roles").Delete(&role); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}
