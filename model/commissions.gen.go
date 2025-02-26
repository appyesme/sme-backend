package model

import (
	"time"
)

const TableNameCommission = "commissions"

type Commission struct {
	ID                   string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt            time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy            string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	CommissionPercentage float64   `gorm:"column:commission_percentage;type:double precision;not null" json:"commission_percentage"`
	GstPercentage        float64   `gorm:"column:gst_percentage;type:double precision;not null" json:"gst_percentage"`
}

func (*Commission) TableName() string {
	return TableNameCommission
}
