package models

import (
	"time"

	"gorm.io/gorm"
)

type PlanType string

const (
	PlanFree     PlanType = "free"
	PlanStandard PlanType = "standard" // Paid
	PlanPremium  PlanType = "premium"  // Paid
)

type Plan struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"` // e.g. "Free Tier", "Gold Monthly"
	Type         PlanType       `gorm:"type:varchar(50);not null" json:"type"`
	Price        float64        `json:"price"`
	MaxUsers     int            `json:"max_users"`     // 0 = Unlimited
	MaxProducts  int            `json:"max_products"`  // 0 = Unlimited
	StorageLimit int            `json:"storage_limit"` // MBs
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
