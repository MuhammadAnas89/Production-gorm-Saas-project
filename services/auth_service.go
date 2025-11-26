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

func (s *AuthService) Login(username, password string) (*models.User, string, error) {
	// Try master DB first
	var masterUser models.User
	err := config.MasterDB.
		Preload("Roles.Permissions").
		Where("username = ?", username).
		First(&masterUser).Error

	if err == nil {
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
		err = tenantDB.
			Preload("Roles.Permissions").
			Where("username = ? AND tenant_id = ?", username, tenant.ID).
			First(&tenantUser).Error

		if err != nil {
			continue
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

func (s *AuthService) ValidateAPIKey(apiKey string) (*models.User, *gorm.DB, error) {
	if !utils.ValidateAPIKeyFormat(apiKey) {
		return nil, nil, errors.New("invalid API key format")
	}

	tenantID, err := utils.ExtractTenantIDFromAPIKey(apiKey)
	if err != nil {
		return nil, nil, err
	}

	var masterUser models.User
	err = config.MasterDB.
		Preload("Roles.Permissions").
		Where("api_key = ?", apiKey).
		First(&masterUser).Error

	if err == nil {
		if !masterUser.IsActive {
			return nil, nil, errors.New("user account is disabled")
		}
		if utils.IsAPIKeyExpired(masterUser.APIKeyExpiry) {
			return nil, nil, errors.New("API key expired")
		}
		return &masterUser, config.MasterDB, nil
	}

	tenant, err := s.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, nil, err
	}

	tenantDB, err := config.TenantManager.GetTenantDB(tenant)
	if err != nil {
		return nil, nil, err
	}

	var tenantUser models.User
	err = tenantDB.
		Preload("Roles.Permissions").
		Where("api_key = ?", apiKey).
		First(&tenantUser).Error

	if err != nil {
		return nil, nil, errors.New("invalid API key")
	}

	if !tenantUser.IsActive {
		return nil, nil, errors.New("user account is disabled")
	}
	if utils.IsAPIKeyExpired(tenantUser.APIKeyExpiry) {
		return nil, nil, errors.New("API key expired")
	}

	return &tenantUser, tenantDB, nil
}

// CreateInitialSuperAdmin function remove kar di gayi hai, kyunki ab SeedMasterData use hoga

func (s *AuthService) Logout(apiKey string) error {
	tenantID, err := utils.ExtractTenantIDFromAPIKey(apiKey)
	if err != nil {
		return err
	}

	// Master DB mein clear karo
	var masterUser models.User
	err = config.MasterDB.Where("api_key = ?", apiKey).First(&masterUser).Error
	if err == nil {
		masterUser.APIKey = ""
		masterUser.APIKeyExpiry = nil
		config.MasterDB.Save(&masterUser)
	}

	// Tenant database mein bhi clear karo
	tenant, err := s.tenantRepo.GetByID(tenantID)
	if err == nil {
		tenantDB, err := config.TenantManager.GetTenantDB(tenant)
		if err == nil {
			var tenantUser models.User
			err = tenantDB.Where("api_key = ?", apiKey).First(&tenantUser).Error
			if err == nil {
				tenantUser.APIKey = ""
				tenantUser.APIKeyExpiry = nil
				tenantDB.Save(&tenantUser)
			}
		}
	}

	return nil
}
