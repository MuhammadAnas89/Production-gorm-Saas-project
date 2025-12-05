package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	Description  string `json:"description"`
	IsSystemRole bool   `gorm:"default:false" json:"is_system_role"`

	TenantID uint `gorm:"index;not null" json:"tenant_id"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	Users       []User       `gorm:"many2many:user_roles;" json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
