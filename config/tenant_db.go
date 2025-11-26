package config

import (
	"fmt"
	"go-multi-tenant/models"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TenantDBManager struct {
	tenantDBs map[uint]*gorm.DB
	mutex     sync.RWMutex
	config    *Config
}

var TenantManager *TenantDBManager

func InitTenantManager(cfg *Config) {
	TenantManager = &TenantDBManager{
		tenantDBs: make(map[uint]*gorm.DB),
		config:    cfg,
	}
}

func (tm *TenantDBManager) GetTenantDB(tenant *models.Tenant) (*gorm.DB, error) {
	if tenant == nil {
		return nil, fmt.Errorf("tenant cannot be nil")
	}

	if tenant.ID == 0 {
		return nil, fmt.Errorf("tenant ID cannot be zero")
	}

	tm.mutex.RLock()
	if db, exists := tm.tenantDBs[tenant.ID]; exists {
		tm.mutex.RUnlock()
		return db, nil
	}
	tm.mutex.RUnlock()

	return tm.initializeTenantDB(tenant)
}

// ✅ NEW: Helper function to sync Global Permissions to Tenant DB
func SyncPermissions(tenantDB *gorm.DB) error {
	// 1. Master DB se Global Modules aur Permissions uthao
	var globalModules []models.Module
	// MasterDB variable same package (config) mein available hai
	if err := MasterDB.Preload("Permissions").Find(&globalModules).Error; err != nil {
		return fmt.Errorf("failed to fetch global modules: %w", err)
	}

	// 2. Tenant DB mein Transaction shuru karo
	return tenantDB.Transaction(func(tx *gorm.DB) error {
		for _, mod := range globalModules {
			// Local Module check karo ya create karo
			var localModule models.Module
			if err := tx.FirstOrCreate(&localModule, models.Module{Name: mod.Name}).Error; err != nil {
				return err
			}

			// Description update karo (sync description changes)
			localModule.Description = mod.Description
			if err := tx.Save(&localModule).Error; err != nil {
				return err
			}

			// Permissions sync karo
			for _, perm := range mod.Permissions {
				var localPerm models.Permission
				if err := tx.Where("name = ?", perm.Name).First(&localPerm).Error; err != nil {
					// Agar nahi mili, to create karo
					newPerm := models.Permission{
						Name:        perm.Name,
						Description: perm.Description,
						Category:    perm.Category,
						ModuleID:    &localModule.ID, // Local Module ID link karo
					}
					if err := tx.Create(&newPerm).Error; err != nil {
						return err
					}
				} else {
					// Update karo (description/category waghaira)
					localPerm.Description = perm.Description
					localPerm.Category = perm.Category
					localPerm.ModuleID = &localModule.ID
					if err := tx.Save(&localPerm).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func (tm *TenantDBManager) initializeTenantDB(tenant *models.Tenant) (*gorm.DB, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Double-check after acquiring write lock
	if db, exists := tm.tenantDBs[tenant.ID]; exists {
		return db, nil
	}

	actualDBName := tenant.GetActualDBName()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		tm.config.DBUser, tm.config.DBPassword, tm.config.DBHost, actualDBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database %s: %w", actualDBName, err)
	}

	// Configure connection pool for tenant DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB for tenant database: %w", err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Module{}, // Module table bhi zaroori hai
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate tenant database %s: %w", actualDBName, err)
	}

	// ✅ NEW: AutoMigrate ke foran baad SyncPermissions call karo
	if err := SyncPermissions(db); err != nil {
		// Log error but don't fail initialization, taaki app chalti rahe
		log.Printf("⚠️ Warning: Failed to sync permissions for tenant %s: %v", tenant.Name, err)
	} else {
		log.Printf("✅ Permissions synced for tenant: %s", tenant.Name)
	}

	tm.tenantDBs[tenant.ID] = db
	log.Printf("Lazy initialized tenant database: %s (Tenant: %s)", actualDBName, tenant.Name)
	return db, nil
}

func (tm *TenantDBManager) CreateDedicatedDatabase(tenant *models.Tenant) error {
	if tenant.DatabaseType != models.DedicatedDB {
		return nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/?charset=utf8mb4&parseTime=True&loc=Local",
		tm.config.DBUser, tm.config.DBPassword, tm.config.DBHost)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL for database creation: %w", err)
	}

	// Close the connection when done
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", tenant.DBName)).Error
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", tenant.DBName, err)
	}

	log.Printf("Created dedicated database: %s", tenant.DBName)
	return nil
}

func (tm *TenantDBManager) CreateSharedDatabase() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/?charset=utf8mb4&parseTime=True&loc=Local",
		tm.config.DBUser, tm.config.DBPassword, tm.config.DBHost)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL for shared database creation: %w", err)
	}

	// Close the connection when done
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	err = db.Exec("CREATE DATABASE IF NOT EXISTS `shared_tenants_db`").Error
	if err != nil {
		return fmt.Errorf("failed to create shared database: %w", err)
	}

	log.Printf("Created/verified shared database: shared_tenants_db")
	return nil
}

func (tm *TenantDBManager) GetCachedTenantCount() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return len(tm.tenantDBs)
}

func (tm *TenantDBManager) ClearCache() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Close all database connections before clearing
	for tenantID, db := range tm.tenantDBs {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		delete(tm.tenantDBs, tenantID)
	}

	log.Println("Tenant database cache cleared and connections closed")
}

func (tm *TenantDBManager) RemoveTenantFromCache(tenantID uint) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if db, exists := tm.tenantDBs[tenantID]; exists {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		delete(tm.tenantDBs, tenantID)
		log.Printf("Removed tenant %d from cache", tenantID)
	}
}

func GetTenantManager() *TenantDBManager {
	return TenantManager
}
