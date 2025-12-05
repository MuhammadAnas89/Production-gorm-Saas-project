package models

import (
	"time"

	"gorm.io/gorm"
)

type GlobalIdentity struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Email    string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	TenantID uint   `gorm:"not null;index" json:"tenant_id"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
