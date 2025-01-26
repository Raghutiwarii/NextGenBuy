package models

import "gorm.io/gorm"

func OmitIDToDeletedAtFields(db *gorm.DB) *gorm.DB {
	return db.Omit("created_at", "updated_at", "deleted_at")
}
