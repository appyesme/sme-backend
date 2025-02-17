package model

import (
	"time"
)

const TableNameServiceMedia = "service_medias"

// ServiceMedia mapped from table <service_medias>
type ServiceMedia struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy   string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	ServiceID   string    `gorm:"column:service_id;type:uuid;not null" json:"service_id"`
	FileName    string    `gorm:"column:file_name;type:text;not null" json:"file_name"`
	URL         string    `gorm:"column:url;type:text;not null" json:"url"`
	StoragePath string    `gorm:"column:storage_path;type:text;not null" json:"storage_path"`
}

// TableName ServiceMedia's table name
func (*ServiceMedia) TableName() string {
	return TableNameServiceMedia
}
