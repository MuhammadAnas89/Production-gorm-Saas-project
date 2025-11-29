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
	// ✅ FIX: Use CacheService instead of config methods
	cacheService := NewCacheService()
	_ = cacheService.Delete(cacheKey)
}

// --- Core Functions ---

func (s *UserService) CreateUser(tenantDB *gorm.DB, tenantID uint, req *CreateUserRequest, currentUser *models.User) (*models.User, error) {
	// 1. Check Permissions
	if !currentUser.HasPermission("user:create") {
		return nil, errors.New("insufficient permissions")
	}

	// 2. ✅ PLAN LIMIT CHECK (Users)
	var tenant models.Tenant
	// ✅ FIX: Use config.GetMasterDB()
	if err := config.GetMasterDB().Preload("Plan").First(&tenant, tenantID).Error; err != nil {
		return nil, errors.New("failed to load tenant info")
	}

	if tenant.Plan.MaxUsers > 0 { // 0 means unlimited
		var currentCount int64
		tenantDB.Model(&models.User{}).Count(&currentCount)

		if int(currentCount) >= tenant.Plan.MaxUsers {
			return nil, fmt.Errorf("plan limit reached: your plan allows max %d users", tenant.Plan.MaxUsers)
		}
	}

	// 3. Check Global Uniqueness
	var count int64
	// ✅ FIX: Use config.GetMasterDB()
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

	// 5. Assign Role
	var role models.Role
	if err := tenantDB.First(&role, req.RoleID).Error; err == nil {
		tenantDB.Model(user).Association("Roles").Append(&role)
	}

	// 6. ✅ REGISTER GLOBAL IDENTITY
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

	// Preload roles manually if needed (Repo usually does it)
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
		// ✅ FIX: Use GetByUsername (Standard Repo method)
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
		// ✅ FIX: Use GetByEmail (Fixed signature)
		existingEmailUser, err := userRepo.GetByEmail(emailStr)
		if err == nil && existingEmailUser != nil && existingEmailUser.ID != userID {
			return nil, errors.New("email already exists")
		}
		user.Email = emailStr
	}

	if isActive, exists := updateData["is_active"]; exists {
		user.IsActive = isActive.(bool)
	}

	// Role Update
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

	// ✅ FIX: Use CacheService logic
	cacheService := NewCacheService()
	cacheKey := fmt.Sprintf("tenant:%d:users:list", currentUser.TenantID)
	var cachedUsers []models.User

	// Try Cache
	err := cacheService.Get(cacheKey, &cachedUsers)
	if err == nil {
		fmt.Println("🔥 REDIS CACHE HIT!")
		return cachedUsers, nil
	}

	fmt.Println("🐢 REDIS MISS!")

	userRepo := repositories.NewUserRepository(tenantDB)
	// ✅ FIX: Use List with 0 offset (Fetch All or paginate as needed)
	// Assuming -1 or 0 fetches all based on your logic, here passing 0, 100 for now or implement full list
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// Set Cache
	_ = cacheService.Set(cacheKey, users, 0) // 0 ttl = default/infinite depending on redis config, or use time.Minute

	return users, nil
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
	// Direct DB access is fine for reading static data like roles
	if err := tenantDB.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	return roles, nil
}

// Assign Role Logic
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
