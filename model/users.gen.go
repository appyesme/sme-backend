package model

import (
	"time"

	"github.com/lib/pq"
)

const TableNameUser = "users"

// User mapped from table <users>
type User struct {
	ID                  string          `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt           time.Time       `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Name                string          `gorm:"column:name;type:text;not null" json:"name"`
	Email               *string         `gorm:"column:email;type:text" json:"email"`
	PhotoURL            *string         `gorm:"column:photo_url;type:text" json:"photo_url"`
	TotalWorkExperience *int16          `gorm:"column:total_work_experience;type:smallint" json:"total_work_experience"`
	Expertises          *pq.StringArray `gorm:"column:expertises;type:text[]" json:"expertises"`
	Documents           *pq.StringArray `gorm:"column:documents;type:text[]" json:"documents"`
	Verified            bool            `gorm:"column:verified;type:boolean;not null" json:"verified"`
	About               *string         `gorm:"column:about;type:text" json:"about"`
	AadharNumber        *string         `gorm:"column:aadhar_number;type:character varying(12);default:NULL" json:"aadhar_number"`
	PanNumber           *string         `gorm:"column:pan_number;type:character varying(10);default:NULL" json:"pan_number"`
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}
