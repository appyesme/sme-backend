package model

import (
	"time"
)

const TableNameBankAccount = "bank_accounts"

type BankAccount struct {
	ID            string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy     string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	AccountName   string    `gorm:"column:account_name;type:text;not null" json:"account_name"`
	AccountNumber string    `gorm:"column:account_number;type:character varying;not null" json:"account_number"`
	IfscCode      string    `gorm:"column:ifsc_code;type:character varying;not null" json:"ifsc_code"`
	Upi           *string   `gorm:"column:upi;type:character varying" json:"upi"`
}

func (*BankAccount) TableName() string {
	return TableNameBankAccount
}
