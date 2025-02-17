package model

import (
	"time"
)

const TableNamePayment = "payments"

// Payment mapped from table <payments>
type Payment struct {
	ID            string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy     string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	AppointmentID string    `gorm:"column:appointment_id;type:uuid;not null" json:"appointment_id"`
	Amount        float64   `gorm:"column:amount;type:numeric(10,2);not null" json:"amount"`
	Currency      string    `gorm:"column:currency;type:character varying(10);default:INR" json:"currency"`
	OrderID       string    `gorm:"column:order_id;type:character varying(255)" json:"order_id"`
	PaymentID     string    `gorm:"column:payment_id;type:character varying(255)" json:"payment_id"`
	Status        string    `gorm:"column:status;type:character varying(50);not null" json:"status"`
	ServiceID     string    `gorm:"column:service_id;type:uuid;not null" json:"service_id"`
}

// TableName Payment's table name
func (*Payment) TableName() string {
	return TableNamePayment
}
