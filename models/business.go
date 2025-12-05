package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	TenantID uint   `gorm:"uniqueIndex:idx_category_tenant;not null" json:"tenant_id"`
	Name     string `gorm:"type:varchar(100);uniqueIndex:idx_category_tenant;not null" json:"name"`

	Description string         `json:"description"`
	Products    []Product      `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Product struct {
	ID uint `gorm:"primaryKey" json:"id"`

	TenantID uint   `gorm:"uniqueIndex:idx_sku_tenant;not null" json:"tenant_id"`
	Name     string `gorm:"type:varchar(200);not null" json:"name"`
	SKU      string `gorm:"type:varchar(100);uniqueIndex:idx_sku_tenant" json:"sku"`

	Price      float64        `json:"price"`
	CostPrice  float64        `json:"cost_price"`
	CategoryID *uint          `json:"category_id"`
	Category   *Category      `json:"category,omitempty"`
	Inventory  *Inventory     `gorm:"foreignKey:ProductID" json:"inventory,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
type Inventory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      uint      `gorm:"index;not null" json:"tenant_id"`
	ProductID     uint      `gorm:"uniqueIndex;not null" json:"product_id"`
	Quantity      int       `gorm:"default:0" json:"quantity"`
	LowStockAlert int       `gorm:"default:10" json:"low_stock_alert"`
	Location      string    `json:"location"`
	UpdatedAt     time.Time `json:"updated_at"`
}
