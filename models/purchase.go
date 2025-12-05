package models

import (
	"time"

	"gorm.io/gorm"
)

// Status Constants
const (
	POPending    = "pending"
	PORejected   = "rejected"
	PODispatched = "dispatched"
	POReceived   = "received"
)

type PurchaseOrder struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"index;not null" json:"tenant_id"`
	ProductID   uint           `gorm:"not null" json:"product_id"`
	Product     *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity    int            `gorm:"not null" json:"quantity"`
	BuyPrice    float64        `gorm:"not null" json:"buy_price"`
	Status      string         `gorm:"type:varchar(20);default:'pending'" json:"status"`
	RequestedBy uint           `json:"requested_by"`
	ApprovedBy  *uint          `json:"approved_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
