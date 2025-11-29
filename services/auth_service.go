package services

import (
	"errors"
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"go-multi-tenant/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	tenantRepo repositories.TenantRepository
}

func NewAuthService(tenantRepo repositories.TenantRepository) *AuthService {
	return &AuthService{
		tenantRepo: tenantRepo,
	}
}

// Login: The "Loop-Free" Implementation
func (s *AuthService) Login(email, password string) (*models.User, string, error) {

	// 1. GLOBAL LOOKUP (Master DB)
	// Check karo ye banda poore system mein kahan exist karta hai
	identity, err := s.tenantRepo.GetGlobalIdentity(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("invalid credentials") // User not found globally
		}
		return nil, "", err
	}

	// 2. Fetch Tenant Info
	tenant, err := s.tenantRepo.GetByID(identity.TenantID)
	if err != nil {
		return nil, "", errors.New("tenant not found")
	}

	if !tenant.IsActive {
		return nil, "", errors.New("company account is suspended")
	}

	// 3. Connect to Specific Tenant DB
	// Ab hamain loop chalane ki zaroorat nahi, seedha sahi DB mil gaya
	tenantDB, err := config.TenantManager.GetTenantDB(tenant)
	if err != nil {
		return nil, "", errors.New("database connection failed")
	}

	// 4. Verify Credentials in Tenant DB
	userRepo := repositories.NewUserRepository(tenantDB)
	user, err := userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, "", errors.New("user account is disabled")
	}

	if !utils.VerifyPassword(password, user.Password) {
		return nil, "", errors.New("invalid credentials")
	}

	// 5. Generate JWT Token (12 Hours)
	roleName := "User"
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Name
	}

	token, err := utils.GenerateToken(user.ID, user.TenantID, user.Email, roleName)
	if err != nil {
		return nil, "", fmt.Errorf("token generation failed: %w", err)
	}

	return user, token, nil
}
