package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"uniqueIndex:idx_email_tenant;not null" json:"tenant_id"`
	Username  string         `gorm:"type:varchar(255);not null" json:"username"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex:idx_email_tenant;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Roles     []Role         `gorm:"many2many:user_roles;" json:"roles"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) HasPermission(permName string) bool {
	for _, role := range u.Roles {

		for _, perm := range role.Permissions {
			if perm.Name == permName {
				return true
			}
		}
	}
	return false
}

func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

func (u *User) GetPermissions() []string {
	permMap := make(map[string]bool)
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			permMap[perm.Name] = true
		}
	}

	var perms []string
	for p := range permMap {
		perms = append(perms, p)
	}
	return perms
}
