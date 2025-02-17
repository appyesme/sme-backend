package model

import (
	"time"
)

const TableNameServiceDay = "service_days"

// ServiceDay mapped from table <service_days>
type ServiceDay struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	ServiceID string    `gorm:"column:service_id;type:uuid;not null" json:"service_id"`
	Day       int16     `gorm:"column:day;type:smallint;not null" json:"day"`
	Enabled   bool      `gorm:"column:enabled;type:boolean;not null" json:"enabled"`
}

// TableName ServiceDay's table name
func (*ServiceDay) TableName() string {
	return TableNameServiceDay
}
