package models

import (
	"time"
)

type Role struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Name         string `gorm:"type:varchar(100);uniqueIndex:idx_name_tenant;not null" json:"name"`
	Description  string `gorm:"type:text" json:"description"`
	IsSystemRole bool   `gorm:"default:false" json:"is_system_role"`

	TenantID uint `gorm:"uniqueIndex:idx_name_tenant;not null" json:"tenant_id"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
