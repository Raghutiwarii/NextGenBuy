package models

import (
	"gorm.io/gorm"
)

type UserRole struct {
	gorm.Model
	Role     int    `json:"role" gorm:"not null"`
	RoleName string `json:"role_name" gorm:"not null"`
}

const (
	Admin    = 1 // Admin role
	Merchant = 2 // Merchant role
	Customer = 3 // Customer role
)

func (ur *UserRole) GetRoleName() string {
	switch ur.Role {
	case Admin:
		return "Admin"
	case Merchant:
		return "Merchant"
	case Customer:
		return "Customer"
	default:
		return "Unknown"
	}
}
