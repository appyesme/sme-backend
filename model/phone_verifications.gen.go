package model

import (
	"time"
)

const TableNamePhoneVerification = "phone_verifications"

// PhoneVerification mapped from table <phone_verifications>
type PhoneVerification struct {
	ID           string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	PhoneNumber  string    `gorm:"column:phone_number;type:character varying;not null" json:"phone_number"`
	OtpCode      string    `gorm:"column:otp_code;type:character varying(6);not null" json:"otp_code"`
	OtpExpiresAt time.Time `gorm:"column:otp_expires_at;type:timestamp without time zone;not null" json:"otp_expires_at"`
}

// TableName PhoneVerification's table name
func (*PhoneVerification) TableName() string {
	return TableNamePhoneVerification
}
