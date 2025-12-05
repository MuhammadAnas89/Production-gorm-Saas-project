package services

import (
	"errors"
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"go-multi-tenant/utils"
	"regexp"

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

// --- Helpers ---
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func clearUserCache(tenantID uint) {
	cacheKey := fmt.Sprintf("tenant:%d:users:list", tenantID)
	// Make sure NewCacheService is available in your services package
	cacheService := NewCacheService()
	_ = cacheService.Delete(cacheKey)
}

// --- Core Functions ---

func (s *UserService) CreateUser(tenantDB *gorm.DB, tenantID uint, req *CreateUserRequest, currentUser *models.User) (*models.User, error) {
	// 1. Check Permissions
	if !currentUser.HasPermission("user:create") {
		return nil, errors.New("insufficient permissions")
	}

	// 2. Plan Limit Check
	var tenant models.Tenant
	if err := config.GetMasterDB().Preload("Plan").First(&tenant, tenantID).Error; err != nil {
		return nil, errors.New("failed to load tenant info")
	}

	if tenant.Plan.MaxUsers > 0 {
		var currentCount int64
		tenantDB.Model(&models.User{}).Count(&currentCount)
		if int(currentCount) >= tenant.Plan.MaxUsers {
			return nil, fmt.Errorf("plan limit reached: your plan allows max %d users", tenant.Plan.MaxUsers)
		}
	}

	// 3. Check Global Uniqueness
	var count int64
	config.GetMasterDB().Model(&models.GlobalIdentity{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, errors.New("email already exists in the system")
	}

	// 4. Create User
	hashedPassword, _ := utils.HashPassword(req.Password)
	user := &models.User{
		TenantID: tenantID,
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		IsActive: true,
	}

	userRepo := repositories.NewUserRepository(tenantDB)
	if err := userRepo.Create(user); err != nil {
		return nil, err
	}

	// 5. Assign Role (Using Repo method now)
	if req.RoleID > 0 {
		if err := userRepo.AssignRole(user.ID, req.RoleID); err != nil {
			// Log error but don't fail user creation necessarily, or handle rollback
			fmt.Printf("Failed to assign role: %v\n", err)
		}
	}

	// 6. Register Global Identity
	config.GetMasterDB().Create(&models.GlobalIdentity{
		Email:    req.Email,
		TenantID: tenantID,
	})

	clearUserCache(tenantID)
	return user, nil
}

func (s *UserService) GetUser(tenantDB *gorm.DB, userID uint, currentUser *models.User) (*models.User, error) {
	userRepo := repositories.NewUserRepository(tenantDB)

	// Access control for Tenant User
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

	return user, nil
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

	// Username update
	if username, exists := updateData["username"]; exists {
		existingUser, err := userRepo.GetByUsername(username.(string))
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("username already exists")
		}
		user.Username = username.(string)
	}

	// Email update
	if email, exists := updateData["email"]; exists {
		emailStr := email.(string)
		if !isValidEmail(emailStr) {
			return nil, errors.New("invalid email format")
		}
		existingEmailUser, err := userRepo.GetByEmail(emailStr)
		if err == nil && existingEmailUser != nil && existingEmailUser.ID != userID {
			return nil, errors.New("email already exists")
		}
		user.Email = emailStr
	}

	if isActive, exists := updateData["is_active"]; exists {
		user.IsActive = isActive.(bool)
	}

	// Role Update Logic (✅ NOW CLEANED)
	if roleID, exists := updateData["role_id"]; exists {
		var rID uint
		switch v := roleID.(type) {
		case float64:
			rID = uint(v)
		case uint:
			rID = v
		case int:
			rID = uint(v)
		}

		// 1. Verify role exists (Business Logic Check)
		_, err := userRepo.GetRoleByID(rID)
		if err != nil {
			return nil, errors.New("role not found")
		}

		// 2. Perform DB Operation via Repo (✅ FIXED)
		// Ab hum direct tenantDB calls nahi kar rahe
		if err := userRepo.ReplaceRole(user.ID, rID); err != nil {
			return nil, fmt.Errorf("failed to update role: %w", err)
		}
	}

	// Update basic fields
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

	cacheService := NewCacheService()
	cacheKey := fmt.Sprintf("tenant:%d:users:list", currentUser.TenantID)
	var cachedUsers []models.User

	// Try Cache
	err := cacheService.Get(cacheKey, &cachedUsers)
	if err == nil {
		// fmt.Println("🔥 REDIS CACHE HIT!")
		return cachedUsers, nil
	}

	// fmt.Println("🐢 REDIS MISS!")
	userRepo := repositories.NewUserRepository(tenantDB)

	// Fetching all users (pass appropriate limits if pagination needed)
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// Set Cache
	_ = cacheService.Set(cacheKey, users, 0)
	return users, nil
}
