package appointments_service

import (
	"fmt"
	"sme-backend/model"
	"sme-backend/src/core/helpers"
	"sme-backend/src/enums/appointment_status"
	"sme-backend/src/enums/notification_actions"
	"sme-backend/src/enums/payment_status"
	"sme-backend/src/enums/razorpay_refund_status"
	"sme-backend/src/services/notification_service"
	"sme-backend/src/services/webhook_service"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAppointmentCharge(tx *gorm.DB, service_id string, home_service_needed *bool) (float64, error) {
	// CAUTION: Don't multiply by 100 here.
	// You'll get an error if you try to multiply by 100.
	var totalAmount float64
	var service model.Service
	if err := tx.Where("id = ?", service_id).First(&service).Error; err != nil {
		return totalAmount, err
	}

	var commission model.Commission
	if err := tx.Find(&commission).Error; err != nil {
		return totalAmount, err
	}

	if home_service_needed != nil && *home_service_needed {
		totalAmount = service.Charge + service.AdditionalCharge
	} else {
		totalAmount = service.Charge
	}

	commissionPercentage := (totalAmount * commission.CommissionPercentage) / 100
	cstPercentage := (totalAmount * commission.GstPercentage) / 100

	totalAmount = totalAmount + commissionPercentage + cstPercentage
	return totalAmount, nil
}

func CreateAppointment(tx *gorm.DB, booked_appointment *model.Appointment, user_id string) error {
	booked_appointment.CreatedBy = user_id
	return tx.Create(&booked_appointment).Error
}

func UpdateAppointmentBookedPaymentStatus(db *gorm.DB, payment_payload webhook_service.RazorpayWebhookPayload) error {
	booked_appointment := model.Appointment{Status: appointment_status.BOOKED}
	appointment_id := payment_payload.Payload.Payment.Entity.Notes.AppointmentID
	user_id := payment_payload.Payload.Payment.Entity.Notes.CreatedBy

	if err := db.Clauses(clause.Returning{}).Where("id = ? AND created_by = ?", appointment_id, user_id).Updates(&booked_appointment).Error; err != nil {
		return err
	}

	order_id := payment_payload.Payload.Payment.Entity.OrderID
	payment_id := payment_payload.Payload.Payment.Entity.ID

	payment := &model.Payment{
		Status:    payment_status.PAID,
		PaymentID: payment_id,
	}

	condition := "appointment_id = ? AND created_by = ? AND order_id = ?"
	if err := db.Where(condition, appointment_id, user_id, order_id).Updates(&payment).Error; err != nil {
		return err
	}

	var service model.Service
	if err := db.Where("id = ?", booked_appointment.ServiceID).First(&service).Error; err != nil {
		return err
	}

	actions := []notification_service.NotificationAction{{
		Resource:   notification_actions.SERVICE,
		ResourceID: booked_appointment.ServiceID,
	}}

	// TO ENTREPRENEUR
	entrepreneur_body := fmt.Sprintf("You have new appointment scheduled on %s", booked_appointment.AppointmentDate.Format("02 Jan, 2006"))
	entrepreneur_notification := model.Notification{
		UserID:  service.CreatedBy,
		Title:   "New appointment",
		Body:    entrepreneur_body,
		Actions: *helpers.ToRawMessage(actions),
	}

	// TO user
	user_body := fmt.Sprintf("Your appointment have been scheduled on %s", booked_appointment.AppointmentDate.Format("02 Jan, 2006"))
	user_notification := model.Notification{
		UserID:  user_id,
		Title:   "Appointment booked",
		Body:    user_body,
		Actions: *helpers.ToRawMessage(actions),
	}

	// TRY TO BOOK ONE APPOINTMENT AND CHECK NOTIFICATION
	notifications := []*model.Notification{&entrepreneur_notification, &user_notification}
	notification_service.CreateNotificaton(db, notifications)
	return nil
}

func UpdateAppointmentRefundPaymentStatus(db *gorm.DB, refund_response webhook_service.RzpWebhookRefundReponse) error {
	fmt.Println("REUNFD WEBHOOK CALLED")
	var payment model.Payment

	appointment_id := refund_response.Notes.AppointmentID
	service_id := refund_response.Notes.ServiceID
	order_id := refund_response.Notes.OrderID
	payment_id := refund_response.Notes.PaymentId
	user_id := refund_response.Notes.CreatedBy

	where_conditions := `appointment_id = ? AND payment_id = ? AND order_id = ? AND created_by = ?`
	if err := db.Where(where_conditions, appointment_id, payment_id, order_id, user_id).First(&payment).Error; err != nil {
		return err
	}

	if payment.Status == payment_status.REFUND_SETTLED {
		// simply return if payment refund was already done.
		return nil
	} else if refund_response.Status == razorpay_refund_status.PROCESSED {
		payment.Status = payment_status.REFUND_SETTLED

		if err := db.Where("id = ?", payment.ID).Updates(&payment).Error; err != nil {
			return err
		}

		var service model.Service
		if err := db.Where("id = ?", service_id).First(&service).Error; err != nil {
			return err
		}

		var user model.User
		if err := db.Where("id = ?", user_id).First(&user).Error; err != nil {
			return err
		}

		// TO ENTREPRENEUR
		entrepreneur_body := fmt.Sprintf("Refund Rs.%v success for the appointment booked by %v for the service %v", payment.Amount, user.Name, service.Title)
		entrepreneur_notification := model.Notification{
			UserID: service.CreatedBy,
			Title:  "Refund success",
			Body:   entrepreneur_body,
			// Actions: *helpers.ToRawMessage(actions),
		}

		// TO user
		user_body := fmt.Sprintf("Refund Rs.%v success for the service %v", payment.Amount, service.Title)
		user_notification := model.Notification{
			UserID: user_id,
			Title:  "Refund success",
			Body:   user_body,
			// Actions: *helpers.ToRawMessage(actions),
		}

		notifications := []*model.Notification{&entrepreneur_notification, &user_notification}
		notification_service.CreateNotificaton(db, notifications)
	}

	return nil
}
