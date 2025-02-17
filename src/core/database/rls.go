package database

import (
	"fmt"

	"gorm.io/gorm"
)

func SetTenant(db *gorm.DB, tenant_id string) {
	db.Exec(fmt.Sprintf("SET request.jwt.claim.tenant_id = \"%s\"", tenant_id))
}

func UnsetTenant(db *gorm.DB) {
	db.Exec("RESET request.jwt.claim.tenant_id")
}

func SetRLS(db *gorm.DB, user_id string) {
	db.Exec(fmt.Sprintf("SET request.jwt.claim.sub = \"%s\"", user_id))
}

func UnsetRLS(db *gorm.DB) {
	db.Exec("RESET request.jwt.claim.sub")
}
