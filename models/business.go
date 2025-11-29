package models

import (
	"time"

	"gorm.io/gorm"
)

// 1. Category
type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"index;not null" json:"tenant_id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `json:"description"`
	Products    []Product      `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// 2. Product
type Product struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	TenantID  uint    `gorm:"index;not null" json:"tenant_id"`
	Name      string  `gorm:"type:varchar(200);not null" json:"name"`
	SKU       string  `gorm:"type:varchar(100);index" json:"sku"` // Stock Keeping Unit
	Price     float64 `json:"price"`
	CostPrice float64 `json:"cost_price"`

	CategoryID *uint     `json:"category_id"`
	Category   *Category `json:"category,omitempty"`

	// Inventory Relation
	Inventory *Inventory `gorm:"foreignKey:ProductID" json:"inventory,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 3. Inventory (Stock)
type Inventory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      uint      `gorm:"index;not null" json:"tenant_id"`
	ProductID     uint      `gorm:"uniqueIndex;not null" json:"product_id"`
	Quantity      int       `gorm:"default:0" json:"quantity"`
	LowStockAlert int       `gorm:"default:10" json:"low_stock_alert"`
	Location      string    `json:"location"` // Warehouse A, Shelf B
	UpdatedAt     time.Time `json:"updated_at"`
}

// 4. Audit Log (For Reporting)
// Reporting ke liye hum alag se actions track karte hain
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"index;not null" json:"tenant_id"`
	UserID    uint      `json:"user_id"`
	Action    string    `json:"action"`                   // e.g., "product_created", "stock_updated"
	Details   string    `gorm:"type:text" json:"details"` // JSON data of change
	CreatedAt time.Time `json:"created_at"`
}
