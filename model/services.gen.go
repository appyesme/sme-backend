package model

import (
	"time"
)

const TableNameService = "services"

// Service mapped from table <services>
type Service struct {
	ID               string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt        time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy        string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	Title            string    `gorm:"column:title;type:text;not null" json:"title"`
	Expertises       string    `gorm:"column:expertises;type:text;not null" json:"expertises"`
	Charge           float64   `gorm:"column:charge;type:double precision;not null" json:"charge"`
	AdditionalCharge float64   `gorm:"column:additional_charge;type:double precision;not null" json:"additional_charge"`
	HomeAvailable    bool      `gorm:"column:home_available;type:boolean;not null" json:"home_available"`
	Description      *string   `gorm:"column:description;type:text" json:"description"`
	Address          *string   `gorm:"column:address;type:text" json:"address"`
	Status           string    `gorm:"column:status;type:text;not null;default:DRAFTED" json:"status"`
	SalonAvailable   bool      `gorm:"column:salon_available;type:boolean;not null" json:"salon_available"`
}

// TableName Service's table name
func (*Service) TableName() string {
	return TableNameService
}
