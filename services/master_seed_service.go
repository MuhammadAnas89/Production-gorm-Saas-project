package services

import (
	"go-multi-tenant/models"
	"go-multi-tenant/utils"
	"log"
	"os"

	"gorm.io/gorm"
)

func SeedMasterData(db *gorm.DB, username, email, password string) error {

	plans := []models.Plan{
		{
			Name: "Free Starter", Type: models.PlanFree,
			Price: 0, MaxUsers: 2, MaxProducts: 5, StorageLimit: 500, IsActive: true,
		},
		{
			Name: "Pro Monthly", Type: models.PlanStandard,
			Price: 29.99, MaxUsers: 10, MaxProducts: 100, StorageLimit: 5000, IsActive: true,
		},
		{
			Name: "Pro Yearly", Type: models.PlanPremium,
			Price: 299.99, MaxUsers: 10, MaxProducts: 100, StorageLimit: 5000, IsActive: true,
		},
	}

	for _, p := range plans {
		if err := db.Where("name = ?", p.Name).FirstOrCreate(&p).Error; err != nil {
			log.Printf("Error seeding plan %s: %v", p.Name, err)
		}
	}

	// Internal Plan for Super Admin (Unlimited)
	systemPlan := models.Plan{
		Name:     "System Unlimited",
		Type:     "system_internal",
		Price:    0,
		MaxUsers: 0, MaxProducts: 0, StorageLimit: 0,
		IsActive: true,
	}
	if err := db.Where("name = ?", systemPlan.Name).FirstOrCreate(&systemPlan).Error; err != nil {
		return err
	}

	masterTenant := models.Tenant{
		Name:         "Master Tenant",
		DatabaseType: models.DedicatedDB,
		DBName:       "master_db",
		IsActive:     true,
		PlanID:       systemPlan.ID,
	}

	if err := db.Where("name = ?", masterTenant.Name).FirstOrCreate(&masterTenant).Error; err != nil {
		return err
	}

	modules := []models.Module{
		{Name: "User Management", Description: "Manage tenant users and roles"},
		{Name: "Product Management", Description: "Manage products catalog"},
		{Name: "Category Management", Description: "Manage product categories"},
		{Name: "Inventory Management", Description: "Track stock and warehouses"},
		{Name: "Reporting", Description: "View sales and audit logs"},
		{Name: "System Admin", Description: "Super Admin only features"},
	}

	for i := range modules {
		if err := db.FirstOrCreate(&modules[i], models.Module{Name: modules[i].Name}).Error; err != nil {
			log.Printf("Error seeding module %s: %v", modules[i].Name, err)
		}
	}

	permissions := []models.Permission{
		{Name: "user:create", Category: "user", ModuleID: &modules[0].ID},
		{Name: "user:read", Category: "user", ModuleID: &modules[0].ID},
		{Name: "user:update", Category: "user", ModuleID: &modules[0].ID},
		{Name: "user:delete", Category: "user", ModuleID: &modules[0].ID},
		{Name: "role:manage", Category: "role", ModuleID: &modules[0].ID},

		{Name: "product:create", Category: "product", ModuleID: &modules[1].ID},
		{Name: "product:read", Category: "product", ModuleID: &modules[1].ID},
		{Name: "product:update", Category: "product", ModuleID: &modules[1].ID},
		{Name: "product:delete", Category: "product", ModuleID: &modules[1].ID},

		{Name: "category:create", Category: "category", ModuleID: &modules[2].ID},
		{Name: "category:read", Category: "category", ModuleID: &modules[2].ID},
		{Name: "category:update", Category: "category", ModuleID: &modules[2].ID},
		{Name: "category:delete", Category: "category", ModuleID: &modules[2].ID},

		{Name: "inventory:read", Category: "inventory", ModuleID: &modules[3].ID},
		{Name: "inventory:update", Category: "inventory", ModuleID: &modules[3].ID},

		{Name: "report:view", Category: "report", ModuleID: &modules[4].ID},

		{Name: "tenant:create", Category: "system", ModuleID: &modules[5].ID},
		{Name: "tenant:manage", Category: "system", ModuleID: &modules[5].ID},
		{Name: "plan:manage", Category: "system", ModuleID: &modules[5].ID},
		{Name: "system:manage", Category: "system", ModuleID: &modules[5].ID},
	}

	for i := range permissions {
		if err := db.FirstOrCreate(&permissions[i], models.Permission{Name: permissions[i].Name}).Error; err != nil {
			log.Printf("Error seeding permission %s: %v", permissions[i].Name, err)
		}
	}

	superAdminRole := models.Role{
		Name:         "Super Administrator",
		Description:  "System Owner - Full Access",
		IsSystemRole: true,
		TenantID:     masterTenant.ID,
	}

	if err := db.Where("name = ? AND tenant_id = ?", superAdminRole.Name, masterTenant.ID).FirstOrCreate(&superAdminRole).Error; err != nil {
		return err
	}

	var allPermissions []models.Permission
	db.Find(&allPermissions)
	db.Model(&superAdminRole).Association("Permissions").Replace(&allPermissions)

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
	}

	hashed, _ := utils.HashPassword(password)

	var admin models.User
	err := db.Where("email = ?", email).First(&admin).Error

	switch err {
	case gorm.ErrRecordNotFound:
		admin = models.User{
			TenantID: masterTenant.ID,
			Username: username,
			Email:    email,
			Password: hashed,
			IsActive: true,
		}
		if createErr := db.Create(&admin).Error; createErr != nil {
			return createErr
		}
		db.Model(&admin).Association("Roles").Append(&superAdminRole)

		db.Create(&models.GlobalIdentity{
			Email:    email,
			TenantID: masterTenant.ID,
		})
		log.Println("✅ Super Admin Created")

	case nil:
		admin.Password = hashed
		admin.IsActive = true
		if saveErr := db.Save(&admin).Error; saveErr != nil {
			return saveErr
		}

		var count int64
		db.Model(&models.GlobalIdentity{}).Where("email = ?", email).Count(&count)
		if count == 0 {
			db.Create(&models.GlobalIdentity{
				Email:    email,
				TenantID: masterTenant.ID,
			})
		}
		log.Printf("Superadmin user '%s' updated", username)

	default:
		return err
	}

	return nil
}
