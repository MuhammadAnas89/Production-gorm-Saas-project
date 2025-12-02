package config

import (
	"fmt"
	"go-multi-tenant/models"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

// GetTenantDB: Cache check karta hai, agar na mile to connect karta hai
func (tm *TenantDBManager) GetTenantDB(tenant *models.Tenant) (*gorm.DB, error) {
	if tenant.ID == 0 {
		return nil, fmt.Errorf("tenant ID cannot be zero")
	}

	// 1. Check Cache
	tm.mutex.RLock()
	if db, exists := tm.tenantDBs[tenant.ID]; exists {
		tm.mutex.RUnlock()
		return db, nil
	}
	tm.mutex.RUnlock()

	// 2. Initialize
	return tm.initializeTenantDB(tenant)
}

func (tm *TenantDBManager) initializeTenantDB(tenant *models.Tenant) (*gorm.DB, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Double check
	if db, exists := tm.tenantDBs[tenant.ID]; exists {
		return db, nil
	}

	actualDBName := tenant.GetActualDBName()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		tm.config.DBUser, tm.config.DBPassword, tm.config.DBHost, actualDBName)

	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{LogLevel: logger.Error},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant db %s: %w", actualDBName, err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// =====================================================
	// ✅ FIX: Migrations Split kar di hain
	// =====================================================

	// 1. System Tables (Ye HAR database mein hone chahiyen)
	// (Users, Roles, Permissions Master DB mein bhi chahiye Super Admin ke liye)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate system tables: %w", err)
	}

	// 2. Business Tables (Ye Master DB mein NAHI banne chahiyen)
	if actualDBName != "master_db" {
		if err := db.AutoMigrate(
			&models.Category{},
			&models.Product{},
			&models.Inventory{},
			// &models.AuditLog{}, // Agar AuditLog model hai to uncomment kar lena
		); err != nil {
			return nil, fmt.Errorf("failed to migrate business tables: %w", err)
		}
	}

	tm.tenantDBs[tenant.ID] = db
	return db, nil
}

// Helper: Database Creation
func (tm *TenantDBManager) CreateDedicatedDatabase(tenant *models.Tenant) error {
	return tm.createDatabase(tenant.DBName)
}

func (tm *TenantDBManager) CreateSharedDatabase() error {
	return tm.createDatabase("shared_tenants_db")
}

func (tm *TenantDBManager) createDatabase(dbName string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/?charset=utf8mb4",
		tm.config.DBUser, tm.config.DBPassword, tm.config.DBHost)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	return db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)).Error
}

func (tm *TenantDBManager) ClearCache() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for _, db := range tm.tenantDBs {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
	tm.tenantDBs = make(map[uint]*gorm.DB)
}

func (tm *TenantDBManager) GetCachedTenantCount() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return len(tm.tenantDBs)
}
