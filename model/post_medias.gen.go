package model

import (
	"time"
)

const TableNamePostMedia = "post_medias"

// PostMedia mapped from table <post_medias>
type PostMedia struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	PostID      string    `gorm:"column:post_id;type:uuid;not null" json:"post_id"`
	CreatedBy   string    `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	FileName    string    `gorm:"column:file_name;type:text;not null" json:"file_name"`
	URL         string    `gorm:"column:url;type:text;not null" json:"url"`
	StoragePath string    `gorm:"column:storage_path;type:text;not null" json:"storage_path"`
}

// TableName PostMedia's table name
func (*PostMedia) TableName() string {
	return TableNamePostMedia
}
