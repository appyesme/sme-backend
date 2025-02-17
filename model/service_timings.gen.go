package model

import (
	"time"
)

const TableNameServiceTiming = "service_timings"

// ServiceTiming mapped from table <service_timings>
type ServiceTiming struct {
	ID            string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy     string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	ServiceDayID  string    `gorm:"column:service_day_id;type:uuid;not null" json:"service_day_id"`
	StartTime     string    `gorm:"column:start_time;type:time without time zone;not null" json:"start_time"`
	EndTime       string    `gorm:"column:end_time;type:time without time zone;not null" json:"end_time"`
	PeoplePerSlot int16     `gorm:"column:people_per_slot;type:smallint;not null" json:"people_per_slot"`
	Enabled       bool      `gorm:"column:enabled;type:boolean;not null" json:"enabled"`
}

// TableName ServiceTiming's table name
func (*ServiceTiming) TableName() string {
	return TableNameServiceTiming
}
