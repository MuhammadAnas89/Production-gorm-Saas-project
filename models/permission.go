package models

type Permission struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string `json:"description"`
	Category    string `gorm:"type:varchar(50)" json:"category"` // e.g., user, product, system
	ModuleID    *uint  `json:"module_id"`
}
