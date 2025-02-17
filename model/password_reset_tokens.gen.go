package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNamePasswordResetToken = "password_reset_tokens"

// PasswordResetToken mapped from table <password_reset_tokens>
type PasswordResetToken struct {
	ID         string         `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time      `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	CreatedBy  string         `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp with time zone" json:"deleted_at"`
	Token      string         `gorm:"column:token;type:character varying;not null" json:"token"`
	ExpiresAt  time.Time      `gorm:"column:expires_at;type:timestamp with time zone;not null" json:"expires_at"`
	RedirectTo string         `gorm:"column:redirect_to;type:character varying;not null" json:"redirect_to"`
}

// TableName PasswordResetToken's table name
func (*PasswordResetToken) TableName() string {
	return TableNamePasswordResetToken
}
