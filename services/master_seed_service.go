package services

import (
	"log"
	"os"

	"go-multi-tenant/models"
	"go-multi-tenant/utils"

	"gorm.io/gorm"
)

func SeedMasterData(db *gorm.DB, username, email, password string) error {

	masterTenant := models.Tenant{
		Name:         "Master Tenant",
		DatabaseType: models.SharedDB,
		DBName:       "master_tenant_db",
		IsActive:     true,
	}

	if err := db.Where("name = ?", masterTenant.Name).FirstOrCreate(&masterTenant).Error; err != nil {
		log.Printf("warning: failed to seed tenant %s: %v", masterTenant.Name, err)
		return err
	}

	modules := []models.Module{
		{Name: "User Management", Description: "Manage users and their permissions"},
		{Name: "Tenant Management", Description: "Manage tenants and their databases"},
		{Name: "Role Management", Description: "Manage roles and permissions"},
		{Name: "System Administration", Description: "System-wide administration"},
	}

	for i := range modules {
		if err := db.FirstOrCreate(&modules[i], models.Module{Name: modules[i].Name}).Error; err != nil {
			log.Printf("warning: failed to seed module %s: %v", modules[i].Name, err)
		}
	}

	permissions := []models.Permission{

		// User Management Permissions
		{Name: "user:list", Description: "List users", Category: "user", ModuleID: &modules[0].ID},
		{Name: "user:create", Description: "Create new users", Category: "user", ModuleID: &modules[0].ID},
		{Name: "user:read", Description: "View specific users", Category: "user", ModuleID: &modules[0].ID},
		{Name: "user:update", Description: "Update users", Category: "user", ModuleID: &modules[0].ID},
		{Name: "user:delete", Description: "Delete users", Category: "user", ModuleID: &modules[0].ID},

		// Tenant Management Permissions (Sirf Super Admin ke liye relevant hain)
		{Name: "tenant:create", Description: "Create new tenants", Category: "tenant", ModuleID: &modules[1].ID},
		{Name: "tenant:read", Description: "View tenants", Category: "tenant", ModuleID: &modules[1].ID},
		{Name: "tenant:update", Description: "Update tenants", Category: "tenant", ModuleID: &modules[1].ID},
		{Name: "tenant:delete", Description: "Delete tenants", Category: "tenant", ModuleID: &modules[1].ID},

		// Role Management Permissions
		{Name: "role:create", Description: "Create new roles", Category: "role", ModuleID: &modules[2].ID},
		{Name: "role:read", Description: "View roles", Category: "role", ModuleID: &modules[2].ID},
		{Name: "role:update", Description: "Update roles", Category: "role", ModuleID: &modules[2].ID},
		{Name: "role:delete", Description: "Delete roles", Category: "role", ModuleID: &modules[2].ID},
		{Name: "permission:manage", Description: "Manage permissions", Category: "role", ModuleID: &modules[2].ID},

		// System Administration Permissions
		{Name: "system:config", Description: "Configure system settings", Category: "system", ModuleID: &modules[3].ID},
		{Name: "system:monitor", Description: "Monitor system health", Category: "system", ModuleID: &modules[3].ID},
		{Name: "system:backup", Description: "Perform system backups", Category: "system", ModuleID: &modules[3].ID},

		// Full Admin Permission
		{Name: "admin:full", Description: "Full administrative access", Category: "admin", ModuleID: &modules[3].ID},
	}

	for i := range permissions {
		if err := db.FirstOrCreate(&permissions[i], models.Permission{Name: permissions[i].Name}).Error; err != nil {
			log.Printf("warning: failed to seed permission %s: %v", permissions[i].Name, err)
		}
	}

	superAdminRole := models.Role{
		Name:         "Super Administrator",
		Description:  "Full system access with all permissions",
		IsSystemRole: true,
		TenantID:     masterTenant.ID,
	}

	if err := db.Where("name = ? AND tenant_id = ?", superAdminRole.Name, masterTenant.ID).
		FirstOrCreate(&superAdminRole).Error; err != nil {
		return err
	}

	var allPermissions []models.Permission
	if err := db.Find(&allPermissions).Error; err != nil {
		return err
	}

	if err := db.Model(&superAdminRole).Association("Permissions").Replace(&allPermissions); err != nil {
		log.Printf("warning: failed to assign all permissions to Super Administrator: %v", err)
	}

	if username == "" {
		username = os.Getenv("SUPERADMIN_USERNAME")
	}
	if email == "" {
		email = os.Getenv("SUPERADMIN_EMAIL")
	}
	if password == "" {
		password = os.Getenv("SUPERADMIN_PASSWORD")
	}

	if username == "" {
		username = "superadmin"
	}
	if email == "" {
		email = "superadmin@system.com"
	}
	if password == "" {
		password = "Admin123!"
		log.Println("warning: using default SUPERADMIN_PASSWORD; set env SUPERADMIN_PASSWORD in production")
	}

	var admin models.User

	err := db.Where("(username = ? OR email = ?) AND tenant_id = ?", username, email, masterTenant.ID).First(&admin).Error

	hashed, errHash := utils.HashPassword(password)
	if errHash != nil {
		return errHash
	}

	if err == gorm.ErrRecordNotFound {
		admin = models.User{
			TenantID: masterTenant.ID,
			Username: username,
			Email:    email,
			Password: hashed,
			IsActive: true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		log.Printf("Superadmin user '%s' created in Master Tenant", username)
	} else if err == nil {

		admin.Password = hashed
		admin.IsActive = true
		if err := db.Save(&admin).Error; err != nil {
			return err
		}
		log.Printf("Superadmin user '%s' updated", username)
	} else {
		return err
	}

	if err := db.Model(&admin).Association("Roles").Append(&superAdminRole); err != nil {
		log.Printf("warning: failed to attach super role to admin: %v", err)
	}

	log.Println("✅ Master data seeded successfully (Clean Architecture)!")
	log.Printf("Super Administrator Credentials:")
	log.Printf("Username: %s", username)
	log.Printf("Email: %s", email)
	log.Printf("Password: %s", password)

	return nil
}
