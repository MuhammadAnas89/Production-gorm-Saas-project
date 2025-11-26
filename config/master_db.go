package config

import (
	"fmt"
	"go-multi-tenant/models"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MasterDB *gorm.DB

func InitMasterDB(cfg *Config) error {
	if MasterDB != nil {
		return nil // Already initialized
	}

	var err error
	MasterDB, err = gorm.Open(mysql.Open(cfg.MasterDBDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to master database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := MasterDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = MasterDB.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Module{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate master database: %w", err)
	}

	log.Println("Master database connected and migrated successfully")
	return nil
}

func GetMasterDB() *gorm.DB {
	return MasterDB
}

func CloseMasterDB() error {
	if MasterDB != nil {
		sqlDB, err := MasterDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
