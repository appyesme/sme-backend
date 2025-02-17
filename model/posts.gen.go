package model

import (
	"time"
)

const TableNamePost = "posts"

// Post mapped from table <posts>
type Post struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy   string     `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	ServiceID   string     `gorm:"column:service_id;type:uuid;not null" json:"service_id"`
	Description string     `gorm:"column:description;type:text;not null" json:"description"`
	Status      string     `gorm:"column:status;type:text;not null;default:DRAFTED" json:"status"`
}

// TableName Post's table name
func (*Post) TableName() string {
	return TableNamePost
}
