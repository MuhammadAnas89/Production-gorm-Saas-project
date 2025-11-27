package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TenantID     uint           `gorm:"not null" json:"tenant_id"`
	Username     string         `gorm:"type:varchar(255);uniqueIndex:idx_username_tenant;not null" json:"username"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex:idx_email_tenant;not null" json:"email"`
	Password     string         `gorm:"type:varchar(255);not null" json:"-"`
	Roles        []Role         `gorm:"many2many:user_roles;" json:"roles"`
	APIKey       string         `gorm:"type:varchar(255);uniqueIndex;default:NULL" json:"-"`
	APIKeyExpiry *time.Time     `gorm:"index;default:NULL" json:"-"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
type UserResponse struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) HasPermission(permissionName string) bool {
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if perm.Name == permissionName {
				return true
			}
		}
	}
	return false
}

// Check if user has role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// Get user permissions
func (u *User) GetPermissions() []string {
	var permissions []string
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			permissions = append(permissions, perm.Name)
		}
	}
	return permissions
}
