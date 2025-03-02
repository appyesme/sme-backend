package database

import (
	"gorm.io/gorm"
)

// Set RLS for the current transaction
func SetRLS(db *gorm.DB, userID string) error {
	return db.Exec("SELECT set_config('request.jwt.claim.sub', ?, true)", userID).Error
}
