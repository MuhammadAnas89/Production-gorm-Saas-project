package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Category    string         `gorm:"type:varchar(50)" json:"category"`
	ModuleID    *uint          `json:"module_id,omitempty"`
	Module      *Module        `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Roles       []Role         `gorm:"many2many:role_permissions;" json:"roles"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
