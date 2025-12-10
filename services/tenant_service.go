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
}

func NewTenantService(tenantRepo repositories.TenantRepository) *TenantService {
	return &TenantService{tenantRepo: tenantRepo}
}

type CreateTenantRequest struct {
	Name          string              `json:"name"`
	DatabaseType  models.DatabaseType `json:"database_type"`
	PlanID        uint                `json:"plan_id"`
	AdminUsername string              `json:"admin_username"`
	AdminEmail    string              `json:"admin_email"`
	AdminPassword string              `json:"admin_password"`
}

func (s *TenantService) CreateTenant(req *CreateTenantRequest) (*models.Tenant, error) {
	if _, err := s.tenantRepo.GetGlobalIdentity(req.AdminEmail); err == nil {
		return nil, fmt.Errorf("email %s is already registered globally", req.AdminEmail)
	}

	dbName := "shared_tenants_db"
	if req.DatabaseType == models.DedicatedDB {
		dbName = fmt.Sprintf("tenant_%s_db", req.Name)
	}
	var planID uint = req.PlanID
	if planID == 0 {
		var freePlan models.Plan
		config.MasterDB.Where("type = ?", models.PlanFree).First(&freePlan)
		planID = freePlan.ID
	}
	apiKey, err := utils.GenerateSecureKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %v", err)
	}
	tenant := &models.Tenant{
		Name:         req.Name,
		DatabaseType: req.DatabaseType,
		DBName:       dbName,
		IsActive:     true,
		PlanID:       planID,
		APIKey:       apiKey,
	}

	tx := config.MasterDB.Begin()
	if err := tx.Create(tenant).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if req.DatabaseType == models.DedicatedDB {
		if err := config.TenantManager.CreateDedicatedDatabase(tenant); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tenantDB, err := config.TenantManager.GetTenantDB(tenant)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	adminRole := models.Role{
		Name:         "Tenant Admin",
		Description:  "Administrator for this workspace",
		IsSystemRole: true,
		TenantID:     tenant.ID,
	}
	tenantDB.Create(&adminRole)

	var allowedPerms []models.Permission

	config.MasterDB.Where("category NOT IN ?", []string{"system", "admin"}).Find(&allowedPerms)

	if len(allowedPerms) > 0 {
		for _, p := range allowedPerms {
			tenantDB.FirstOrCreate(&models.Permission{Name: p.Name}, p)
		}
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
		tx.Rollback()
		return nil, err
	}
	tenantDB.Model(adminUser).Association("Roles").Append(&adminRole)
	globalID := models.GlobalIdentity{
		Email:    req.AdminEmail,
		TenantID: tenant.ID,
	}
	if err := tx.Create(&globalID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return tenant, nil
}

func (s *TenantService) ListTenants() ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := config.MasterDB.Preload("Plan").Find(&tenants).Error
	return tenants, err
}
