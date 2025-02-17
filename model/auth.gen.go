package model

import (
	"time"
)

const TableNameAuth = "auth"

// Auth mapped from table <auth>
type Auth struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	PhoneNumber string     `gorm:"column:phone_number;type:character varying;not null" json:"phone_number"`
	UserType    string     `gorm:"column:user_type;type:text;not null" json:"user_type"` // Change to Role
	VerifiedAt  *time.Time `gorm:"column:verified_at;type:timestamp with time zone" json:"verified_at"`
}

// TableName Auth's table name
func (*Auth) TableName() string {
	return TableNameAuth
}
