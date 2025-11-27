package services

import (
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"go-multi-tenant/utils"
)

type TenantService struct {
	tenantRepo repositories.TenantRepository
	userRepo   repositories.UserRepository
}

func NewTenantService(tenantRepo repositories.TenantRepository, userRepo repositories.UserRepository) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

type CreateTenantRequest struct {
	Name          string              `json:"name"`
	DatabaseType  models.DatabaseType `json:"database_type"`
	AdminUsername string              `json:"admin_username"`
	AdminEmail    string              `json:"admin_email"`
	AdminPassword string              `json:"admin_password"`
}

type CreateTenantResponse struct {
	Tenant    *models.Tenant `json:"tenant"`
	AdminUser *models.User   `json:"admin_user"`
}

func (s *TenantService) CreateTenant(req *CreateTenantRequest) (*CreateTenantResponse, error) {

	var dbName string
	if req.DatabaseType == models.DedicatedDB {
		dbName = fmt.Sprintf("tenant_%s_db", req.Name)
	} else {
		dbName = "shared_tenants_db"
	}

	tenant := &models.Tenant{
		Name:         req.Name,
		DatabaseType: req.DatabaseType,
		DBName:       dbName,
		IsActive:     true,
	}

	if err := s.tenantRepo.Create(tenant); err != nil {
		return nil, err
	}

	if tenant.DatabaseType == models.DedicatedDB {
		if err := config.TenantManager.CreateDedicatedDatabase(tenant); err != nil {
			s.tenantRepo.Delete(tenant.ID) // Rollback
			return nil, fmt.Errorf("failed to create dedicated database: %w", err)
		}
	} else {
		if err := config.TenantManager.CreateSharedDatabase(); err != nil {
			s.tenantRepo.Delete(tenant.ID) // Rollback
			return nil, fmt.Errorf("failed to create shared database: %w", err)
		}
	}

	tenantDB, err := config.TenantManager.GetTenantDB(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to new tenant db: %w", err)
	}

	// 5. "Tenant Admin" Role Create karo
	adminRole := models.Role{
		Name:         "Tenant Admin",
		Description:  "Administrator for this workspace",
		IsSystemRole: true,
		TenantID:     tenant.ID,
	}

	if err := tenantDB.Where("name = ?", "Tenant Admin").FirstOrCreate(&adminRole).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin role: %w", err)
	}

	var allowedPerms []models.Permission

	if err := tenantDB.Where("category IN ?", []string{"user", "role"}).Find(&allowedPerms).Error; err == nil {
		tenantDB.Model(&adminRole).Association("Permissions").Replace(&allowedPerms)
	}

	hashedPassword, _ := utils.HashPassword(req.AdminPassword)
	adminUser := &models.User{
		TenantID: tenant.ID,
		Username: req.AdminUsername,
		Email:    req.AdminEmail,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := tenantDB.Create(adminUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	if err := tenantDB.Model(adminUser).Association("Roles").Append(&adminRole); err != nil {
		return nil, fmt.Errorf("failed to assign role to user: %w", err)
	}

	return &CreateTenantResponse{
		Tenant:    tenant,
		AdminUser: adminUser,
	}, nil
}

func (s *TenantService) ListTenants() ([]models.Tenant, error) {
	return s.tenantRepo.List()
}

func (s *TenantService) GetTenantByID(id uint) (*models.Tenant, error) {
	return s.tenantRepo.GetByID(id)
}
