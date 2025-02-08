package models

import (
	"time"

	"gorm.io/gorm"
)

type Offer struct {
	gorm.Model
	UUID        string    `gorm:"unique" json:"uuid,omitempty"`
	ProductID   uint      `json:"product_id" gorm:"index"`
	Discount    float64   `json:"discount" gorm:"not null"` // Discount percentage
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active" gorm:"default:false"`
	// Product     Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}
