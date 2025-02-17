package model

import (
	"encoding/json"
	"time"
)

const TableNameNotification = "notifications"

// Notification mapped from table <notifications>
type Notification struct {
	ID        string          `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time       `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	UserID    string          `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	Actions   json.RawMessage `gorm:"column:actions;type:jsonb;not null;default:[]" json:"actions"`
	Title     string          `gorm:"column:title;type:character varying;not null" json:"title"`
	Body      string          `gorm:"column:body;type:text;not null" json:"body"`
	Read      bool            `gorm:"column:read;type:boolean;not null" json:"read"`
	Visible   bool            `gorm:"column:visible;type:boolean;not null;default:true" json:"visible"`
	FcmStatus string          `gorm:"column:fcm_status;type:character varying;not null;default:pending" json:"fcm_status"`
}

// TableName Notification's table name
func (*Notification) TableName() string {
	return TableNameNotification
}
