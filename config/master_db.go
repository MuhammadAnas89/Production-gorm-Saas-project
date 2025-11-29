package config

import (
	"fmt"
	"go-multi-tenant/models"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var MasterDB *gorm.DB

func InitMasterDB(cfg *Config) error {
	var err error

	// Logger taaki errors nazar ayen
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	MasterDB, err = gorm.Open(mysql.Open(cfg.MasterDBDSN), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to master database: %w", err)
	}

	// Connection Pool Settings
	sqlDB, _ := MasterDB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = MasterDB.AutoMigrate(
		&models.GlobalIdentity{},
		&models.Plan{},
		&models.Tenant{},
		&models.Module{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate master database: %w", err)
	}

	log.Println("✅ Master database connected and migrated successfully")
	return nil
}

// Getter Function (Best Practice)
func GetMasterDB() *gorm.DB {
	return MasterDB
}
