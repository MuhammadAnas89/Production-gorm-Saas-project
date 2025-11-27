package models

import "gorm.io/gorm"

type Module struct {
	gorm.Model
	Name        string       `json:"name" gorm:"unique;not null"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `gorm:"foreignKey:ModuleID" json:"permissions,omitempty"`
}
