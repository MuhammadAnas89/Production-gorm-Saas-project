package services

import (
	"errors"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"go-multi-tenant/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   repositories.UserRepository
	tenantRepo repositories.TenantRepository
}

func NewAuthService(userRepo repositories.UserRepository, tenantRepo repositories.TenantRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// Login Function (Logs Cleaned)
func (s *AuthService) Login(username, password string) (*models.User, string, error) {
	// 1. Try Master DB first (For Super Admin)
	var masterUser models.User
	// Use Find() + Limit(1) to prevent 'record not found' error logs
	config.MasterDB.Preload("Roles.Permissions").Where("username = ?", username).Limit(1).Find(&masterUser)

	if masterUser.ID != 0 {
		if !utils.VerifyPassword(password, masterUser.Password) {
			return nil, "", errors.New("invalid credentials")
		}
		if !masterUser.IsActive {
			return nil, "", errors.New("user account is disabled")
		}

		apiKey, err := utils.GenerateAPIKey(masterUser.TenantID)
		if err != nil {
			return nil, "", err
		}

		apiKeyExpiry := utils.GetAPIKeyExpiry()
		masterUser.APIKey = apiKey
		masterUser.APIKeyExpiry = &apiKeyExpiry

		if err := s.userRepo.Update(&masterUser); err != nil {
			return nil, "", err
		}

		return &masterUser, apiKey, nil
	}

	// 2. If not in Master DB, check Tenants
	tenants, err := s.tenantRepo.List()
	if err != nil {
		return nil, "", err
	}

	for _, tenant := range tenants {
		if tenant.Name == "Master Tenant" {
			continue
		}

		tenantDB, err := config.TenantManager.GetTenantDB(&tenant)
		if err != nil {
			continue
		}

		var tenantUser models.User
		query := tenantDB.Preload("Roles.Permissions").Where("username = ?", username)

		// Shared DB Scope
		if tenant.DatabaseType == models.SharedDB {
			query = query.Where("tenant_id = ?", tenant.ID)
		}

		// Use Find() to suppress logs
		query.Limit(1).Find(&tenantUser)

		if tenantUser.ID == 0 {
			continue // User not found here, try next silently
		}

		if !utils.VerifyPassword(password, tenantUser.Password) {
			return nil, "", errors.New("invalid credentials")
		}

		if !tenantUser.IsActive {
			return nil, "", errors.New("user account is disabled")
		}

		apiKey, err := utils.GenerateAPIKey(tenantUser.TenantID)
		if err != nil {
			return nil, "", err
		}

		apiKeyExpiry := utils.GetAPIKeyExpiry()
		tenantUser.APIKey = apiKey
		tenantUser.APIKeyExpiry = &apiKeyExpiry

		if err := tenantDB.Save(&tenantUser).Error; err != nil {
			return nil, "", err
		}

		return &tenantUser, apiKey, nil
	}

	return nil, "", errors.New("invalid credentials")
}

// ValidateAPIKey (Smart Lookup)
func (s *AuthService) ValidateAPIKey(apiKey string) (*models.User, *gorm.DB, error) {
	if !utils.ValidateAPIKeyFormat(apiKey) {
		return nil, nil, errors.New("invalid API key format")
	}

	// 1. Extract Tenant ID (Fastest check)
	tenantID, err := utils.ExtractTenantIDFromAPIKey(apiKey)
	if err != nil {
		return nil, nil, err
	}

	// 2. Get Tenant Info
	var tenant models.Tenant
	if err := config.MasterDB.First(&tenant, tenantID).Error; err != nil {
		return nil, nil, errors.New("invalid tenant in API key")
	}

	// 3. Pick Database
	var targetDB *gorm.DB
	if tenant.Name == "Master Tenant" {
		targetDB = config.MasterDB
	} else {
		targetDB, err = config.TenantManager.GetTenantDB(&tenant)
		if err != nil {
			return nil, nil, errors.New("database connection failed")
		}
	}

	// 4. Single DB Query
	var user models.User
	query := targetDB.Preload("Roles.Permissions").Where("api_key = ?", apiKey)

	if tenant.DatabaseType == models.SharedDB && tenant.Name != "Master Tenant" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&user).Error; err != nil {
		return nil, nil, errors.New("invalid API key")
	}

	if !user.IsActive {
		return nil, nil, errors.New("user account is disabled")
	}
	if utils.IsAPIKeyExpired(user.APIKeyExpiry) {
		return nil, nil, errors.New("API key expired")
	}

	return &user, targetDB, nil
}

func (s *AuthService) Logout(apiKey string) error {
	tenantID, err := utils.ExtractTenantIDFromAPIKey(apiKey)
	if err != nil {
		return err
	}

	var tenant models.Tenant
	if err := config.MasterDB.First(&tenant, tenantID).Error; err != nil {
		return err
	}

	var targetDB *gorm.DB
	if tenant.Name == "Master Tenant" {
		targetDB = config.MasterDB
	} else {
		targetDB, _ = config.TenantManager.GetTenantDB(&tenant)
	}

	if targetDB != nil {
		var user models.User
		if err := targetDB.Where("api_key = ?", apiKey).First(&user).Error; err == nil {
			user.APIKey = ""
			user.APIKeyExpiry = nil
			targetDB.Save(&user)
		}
	}
	return nil
}
