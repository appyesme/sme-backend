package model

import (
	"time"
)

const TableNameAppointment = "appointments"

type Appointment struct {
	ID                string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy         string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	ServiceID         string    `gorm:"column:service_id;type:uuid;not null" json:"service_id"`
	ServiceTimingID   string    `gorm:"column:service_timing_id;type:uuid;not null" json:"service_timing_id"`
	AppointmentDate   time.Time `gorm:"column:appointment_date;type:date;not null" json:"appointment_date"`
	StartTime         string    `gorm:"column:start_time;type:time without time zone;not null" json:"start_time"`
	EndTime           string    `gorm:"column:end_time;type:time without time zone;not null" json:"end_time"`
	Status            string    `gorm:"column:status;type:text;not null;default:INITIATED" json:"status"`
	HomeServiceNeeded bool      `gorm:"column:home_service_needed;type:boolean;not null" json:"home_service_needed"`
	HomeReachTime     *string   `gorm:"column:home_reach_time;type:time without time zone" json:"home_reach_time"`
	HomeAddress       *string   `gorm:"column:home_address;type:text" json:"home_address"`
}

func (*Appointment) TableName() string {
	return TableNameAppointment
}
