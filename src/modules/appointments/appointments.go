package appointments_handler

import (
	"errors"
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/enums/appointment_status"
	"sme-backend/src/enums/payment_status"
	"sme-backend/src/enums/user_types"
	"sme-backend/src/services/appointments_service"
	"sme-backend/src/services/razorpay_servce"
	"strings"
)

func GetAppointmentPrice(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_id := helpers.GetQueryParameter(request, "service_id")
	home_service_needed := helpers.GetQueryBoolParam(request, "home_service_needed")

	if service_id == "" {
		err_msg := "Service id not specified"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	total_amount, err := appointments_service.GetAppointmentCharge(db, service_id, home_service_needed)

	if err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	resp := map[string]any{"amount": total_amount}
	helpers.HandleSuccess(response, http.StatusOK, "Service total price fetched", resp)
}

func BookAppointment(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)

	var booked_appointment model.Appointment
	if err := helpers.ParseBody(request, &booked_appointment, false); err != nil {
		helpers.HandleError(response, http.StatusNotFound, "Provide valid inputs", err)
		return
	}

	var payment model.Payment

	// START TRANSACTION
	tx := db.Begin()
	if err := tx.Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	defer tx.Rollback()

	var registration_fees float64
	registration_fees, err := appointments_service.GetAppointmentCharge(tx, booked_appointment.ServiceID, &booked_appointment.HomeServiceNeeded)

	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	} else if registration_fees == 0 { // FREE
		booked_appointment.Status = appointment_status.BOOKED
		if err := appointments_service.CreateAppointment(tx, &booked_appointment, user_id); err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Unable to book appointment", err)
			return
		}
	} else if registration_fees > 0 {
		if err := appointments_service.CreateAppointment(tx, &booked_appointment, user_id); err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Unable to book appointment", err)
			return
		}

		order_result, err := razorpay_servce.CreateRazorPayOrderForServicePrice(
			registration_fees,
			booked_appointment.ServiceID,
			booked_appointment.ID,
			user_id,
		)

		if err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
			return
		}

		registration_fees = registration_fees * 100

		payment.CreatedBy = user_id
		payment.Status = payment_status.PENDING
		payment.AppointmentID = booked_appointment.ID
		payment.ServiceID = booked_appointment.ServiceID
		payment.Amount = registration_fees
		payment.OrderID = order_result["id"].(string)
		if err := tx.Create(&payment).Error; err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Unable to book appointment", err)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	// END TRANSACTION

	resp := AppointmentInitiatedDto{
		AppointmentID: booked_appointment.ID,
		ServiceID:     booked_appointment.ServiceID,
		Amount:        payment.Amount,
		OrderID:       payment.OrderID,
		Currency:      "INR",
		Status:        booked_appointment.Status,
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Appointment booking initiated", resp)
}

func AccpetAppointment(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_type := helpers.GetUserType(request)
	appointment_id := helpers.GetUrlParam(request, "appointment_id")

	if user_type == user_types.USER {
		err_msg := "You're not authorized to this action"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New("invalid request"))
		return
	}

	var appointment model.Appointment
	if err := db.Where("id = ?", appointment_id).First(&appointment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to fetch appointment", err)
		return
	}

	if appointment.Status == appointment_status.ACCEPTED {
		err_msg := "Appointment already accepted"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	} else if appointment.Status != appointment_status.BOOKED {
		err_msg := "Appointment is not booked or invalid request"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	appointment.Status = appointment_status.ACCEPTED
	if err := db.Where("id = ?", appointment_id).Updates(&appointment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to complete appointment", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Appointment completed", appointment)
}

func RejectAppointment(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_type := helpers.GetUserType(request)
	appointment_id := helpers.GetUrlParam(request, "appointment_id")

	if user_type == user_types.USER {
		err_msg := "You're not authorized to this action"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New("invalid request"))
		return
	}

	tx := db.Begin()
	if err := tx.Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	defer tx.Rollback()

	var appointment model.Appointment
	if err := tx.Where("id = ?", appointment_id).First(&appointment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong while rejecting appointment", err)
		return
	}

	if appointment.Status != appointment_status.BOOKED {
		err_msg := "Appointment is not booked"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	// Check payment is made
	var payment model.Payment
	if err := tx.Where("appointment_id = ? AND created_by = ?", appointment_id, appointment.CreatedBy).First(&payment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong while rejecting appointment", err)
		return
	}

	if payment.Status == payment_status.REFUND_INITIATED {
		err_msg := "Refund is already initiated. Refund will take 5-7 business days to settle."
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	} else if payment.Status == payment_status.REFUND_SETTLED {
		err_msg := "Refund is already settled."
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	} else if payment.Status != payment_status.PAID {
		err_msg := "Payment is not made or refund is not initiated or contact support team."
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	appointment.Status = appointment_status.REJECTED
	if err := tx.Where("id = ?", appointment_id).Updates(&appointment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong while rejecting appointment", err)
		return
	}

	// Initiate refund
	refund_response, err := razorpay_servce.InitiatedInstantPaymentRefund(payment)
	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong while rejecting appointment", err)
		return
	}

	payment.Status = refund_response.PaymentStatus
	payment.RefundID = refund_response.RefundID
	if err := tx.Where("id = ?", payment.ID).Updates(&payment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong while rejecting appointment", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong while rejecting appointment", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Appointment amount refund initiated", appointment)
}

func MarkAsCompleted(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_type := helpers.GetUserType(request)
	appointment_id := helpers.GetUrlParam(request, "appointment_id")

	if user_type == user_types.USER {
		err_msg := "You're not authorized to this action"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New("invalid request"))
		return
	}

	var appointment model.Appointment
	if err := db.Where("id = ?", appointment_id).First(&appointment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to fetch appointment", err)
		return
	}

	if appointment.Status != appointment_status.ACCEPTED {
		err_msg := "Appointment is not accepted or booked"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	appointment.Status = appointment_status.COMPLETED
	if err := db.Where("id = ?", appointment_id).Updates(&appointment).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to complete appointment", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Appointment completed", appointment)
}

func GetAppointmentsEnabledDays(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_id := helpers.GetQueryParameter(request, "service_id")

	if service_id == "" {
		err := errors.New("sercive id is required")
		helpers.HandleError(response, http.StatusInternalServerError, err.Error(), err)
		return
	}

	var services_days []model.ServiceDay
	query := `select sd.* from
			  service_days sd
			  join service_timings st on st.service_day_id = sd.id and st.people_per_slot != 0 and st.enabled
			  where sd.service_id = ? group by sd.id`

	if err := db.Raw(query, service_id).Scan(&services_days).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get service days", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service days fetched", services_days)
}

func GetAppointmentsEnabledDayTimings(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_day_id := helpers.GetQueryParameter(request, "service_day_id")
	date := helpers.GetQueryParameter(request, "date") // 2024-10-21

	if service_day_id == "" || date == "" {
		err := errors.New("sercive day id and date are required")
		helpers.HandleError(response, http.StatusInternalServerError, err.Error(), err)
		return
	}

	var services_day_timings []GetAppointmentsEnabledDayTimingsDto
	query := `select
				st.*, st.people_per_slot - coalesce(ap.appointment_count, 0) as remaining_slots
			from
				service_timings st
				left join (
					select service_timing_id, count(id) as appointment_count
					from appointments
					where appointment_date = ?
					group by service_timing_id
				) ap on ap.service_timing_id = st.id
			where st.service_day_id = ? and st.enabled and coalesce(ap.appointment_count, 0) < st.people_per_slot
			group by st.id, ap.appointment_count;`

	if err := db.Raw(query, date, service_day_id).Scan(&services_day_timings).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get service day timings", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service day timings fetched", services_day_timings)
}
