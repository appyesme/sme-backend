package model

import (
	"time"
)

const TableNameLastPaymentCleared = "last_payment_cleared"

type LastPaymentCleared struct {
	ID             string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy      string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	EntrepreneurID string    `gorm:"column:entrepreneur_id;type:uuid;not null" json:"entrepreneur_id"`
}

func (*LastPaymentCleared) TableName() string {
	return TableNameLastPaymentCleared
}
