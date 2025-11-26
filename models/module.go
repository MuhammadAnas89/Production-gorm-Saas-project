package models

import "gorm.io/gorm"

// Module represents a grouping for permissions (e.g. users, tenants, billing)
type Module struct {
	gorm.Model
	Name        string       `json:"name" gorm:"unique;not null"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `gorm:"foreignKey:ModuleID" json:"permissions,omitempty"`
}
