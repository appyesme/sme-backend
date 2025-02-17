package notification_service

import (
	"sme-backend/model"

	"gorm.io/gorm"
)

type NotificationAction struct {
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
}

func CreateNotificaton(tx *gorm.DB, notifications []*model.Notification) error {
	return tx.Create(&notifications).Error
}

func GetUserNotificatons(tx *gorm.DB, page, limit int, user_id string, notifications *[]model.Notification) error {
	return tx.Where("user_id = ?", user_id).Order("created_at DESC").Limit(limit).Offset(page * limit).Find(&notifications).Error
}

func MarkAsRead(tx *gorm.DB, noitifiction_id, user_id string) error {
	notification := map[string]interface{}{"read": true}
	return tx.Model(model.Notification{}).Where("id = ?", noitifiction_id).Updates(&notification).Error
}

func UnReadCount(tx *gorm.DB, count *int64, userID string) error {
	return tx.Model(&model.Notification{}).Where("user_id = ? AND read = false", userID).Count(count).Error
}
