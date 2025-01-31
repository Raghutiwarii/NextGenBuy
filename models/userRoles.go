package models

import (
	"gorm.io/gorm"
)

type RoleID uint64

type userRoleRepo struct {
	db *gorm.DB
}

type UserRole struct {
	gorm.Model
	Role     int    `json:"role" gorm:"not null"`
	RoleName string `json:"role_name" gorm:"not null"`
}

const (
	AdminRole    = 1 // Admin role
	MerchantRole = 2 // Merchant role
	CustomerRole = 3 // Customer role
)

func (ur *UserRole) GetRoleName() string {
	switch ur.Role {
	case AdminRole:
		return "Admin"
	case MerchantRole:
		return "Merchant"
	case CustomerRole:
		return "Customer"
	default:
		return "Unknown"
	}
}
