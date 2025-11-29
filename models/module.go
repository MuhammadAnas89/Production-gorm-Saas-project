package models

type Module struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique;not null" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"foreignKey:ModuleID" json:"permissions,omitempty"`
}
