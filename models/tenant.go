package models

import (
	"time"

	"gorm.io/gorm"
)

type DatabaseType string

const (
	SharedDB    DatabaseType = "shared"
	DedicatedDB DatabaseType = "dedicated"
)

type Tenant struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	DatabaseType DatabaseType   `gorm:"type:enum('shared','dedicated');not null" json:"database_type"`
	DBName       string         `gorm:"type:varchar(255);not null" json:"db_name"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *Tenant) GetActualDBName() string {
	if t.DatabaseType == SharedDB {
		return "shared_tenants_db"
	}
	return t.DBName
}
