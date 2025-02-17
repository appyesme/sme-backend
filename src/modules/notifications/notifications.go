package notifications

import (
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/services/notification_service"
)

func GetUserNotifications(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)
	page := helpers.GetQueryPageParam(request)
	limit := helpers.GetQueryLimitParam(request)

	var notifications []model.Notification
	if err := notification_service.GetUserNotificatons(db, page, limit, user_id, &notifications); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get notifications", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Notifications fetched", notifications)
}

func MarkAsRead(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)
	noitifiction_id := helpers.GetUrlParam(request, "noitifiction_id")
	if err := notification_service.MarkAsRead(db, noitifiction_id, user_id); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to make notifications mark as read", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Mark as read", nil)
}

func UnreadCount(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)

	var count int64
	if err := notification_service.UnReadCount(db, &count, user_id); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get unread count", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Unread count fetched", count)
}
