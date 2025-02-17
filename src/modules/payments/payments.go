package payments

import (
	"errors"
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/enums/appointment_status"
	"sme-backend/src/enums/payment_status"
	"sme-backend/src/services/razorpay_servce"
	"strings"
)

func VerifyPaymentStatus(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	appointment_id := helpers.GetQueryParameter(request, "appointment_id")
	user_id := helpers.GetUserIdFromJwtToken(request)

	if appointment_id == "" {
		err_msg := "Appointment id not specified in query"
		helpers.HandleError(response, http.StatusInternalServerError, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	var booked_appointment model.Appointment
	if err := db.Where("id = ? AND created_by = ?", appointment_id, user_id).First(&booked_appointment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	var payment model.Payment
	if err := db.Where("appointment_id = ? AND created_by = ?", appointment_id, user_id).First(&payment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	payment_verification := map[string]interface{}{}

	if booked_appointment.Status == appointment_status.BOOKED && payment.Status == payment_status.PAID {
		payment_verification["status"] = payment_status.PAID
	} else {
		is_paid, err := razorpay_servce.VerifyPaymentByOrderID(payment.OrderID)

		if !is_paid {
			payment_verification["status"] = payment_status.PENDING
		} else {
			// START TRANSACTION
			tx := db.Begin()
			if err := tx.Error; err != nil {
				helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
				return
			}

			defer tx.Rollback()

			if err != nil {
				helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
				return
			}

			booked_appointment := model.Appointment{Status: appointment_status.BOOKED}
			if err := db.Where("id = ? AND created_by = ?", appointment_id, user_id).Updates(&booked_appointment).Error; err != nil {
				helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
				return
			}

			payment := model.Payment{Status: payment_status.PAID}
			if err := db.Where("appointment_id = ? AND created_by = ?", appointment_id, user_id).Updates(&payment).Error; err != nil {
				helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
				return
			}

			if err := tx.Commit().Error; err != nil {
				helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
				return
			}

			payment_verification["status"] = payment.Status
			// END TRANSACTION
		}
	}

	helpers.HandleSuccess(response, http.StatusOK, "Payment verified status fetched", payment_verification)
}
