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
	PlanID        uint                `json:"plan_id"` // Optional (Default to Free)
	AdminUsername string              `json:"admin_username"`
	AdminEmail    string              `json:"admin_email"`
	AdminPassword string              `json:"admin_password"`
}

func (s *TenantService) CreateTenant(req *CreateTenantRequest) (*models.Tenant, error) {
	// 1. Check Global Email Uniqueness
	if _, err := s.tenantRepo.GetGlobalIdentity(req.AdminEmail); err == nil {
		return nil, fmt.Errorf("email %s is already registered globally", req.AdminEmail)
	}

	// 2. Determine DB Name
	dbName := "shared_tenants_db"
	if req.DatabaseType == models.DedicatedDB {
		dbName = fmt.Sprintf("tenant_%s_db", req.Name)
	}

	// 3. Select Plan
	var planID uint = req.PlanID
	if planID == 0 {
		// Default to Free Plan
		var freePlan models.Plan
		config.MasterDB.Where("type = ?", models.PlanFree).First(&freePlan)
		planID = freePlan.ID
	}

	// 4. Create Tenant Record (Master DB)
	tenant := &models.Tenant{
		Name:         req.Name,
		DatabaseType: req.DatabaseType,
		DBName:       dbName,
		IsActive:     true,
		PlanID:       planID,
	}

	tx := config.MasterDB.Begin()
	if err := tx.Create(tenant).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5. Setup Database
	if req.DatabaseType == models.DedicatedDB {
		if err := config.TenantManager.CreateDedicatedDatabase(tenant); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 6. Connect to Tenant DB
	tenantDB, err := config.TenantManager.GetTenantDB(tenant)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 7. Create "Tenant Admin" Role
	adminRole := models.Role{
		Name:         "Tenant Admin",
		Description:  "Administrator for this workspace",
		IsSystemRole: true,
		TenantID:     tenant.ID,
	}
	tenantDB.Create(&adminRole)

	// 8. Assign Permissions (Only Business Logic, No System/SuperAdmin perms)
	var allowedPerms []models.Permission
	// Filter out "system" category from Master DB
	config.MasterDB.Where("category NOT IN ?", []string{"system", "admin"}).Find(&allowedPerms)

	if len(allowedPerms) > 0 {
		// Sync to Tenant DB
		for _, p := range allowedPerms {
			tenantDB.FirstOrCreate(&models.Permission{Name: p.Name}, p)
		}
		// Assign to Role
		tenantDB.Model(&adminRole).Association("Permissions").Replace(&allowedPerms)
	}

	// 9. Create Admin User in Tenant DB
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

	// 10. ✅ REGISTER GLOBAL IDENTITY (Loop Fix)
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
	// Simple wrapper around repo
	var tenants []models.Tenant
	err := config.MasterDB.Preload("Plan").Find(&tenants).Error
	return tenants, err
}
