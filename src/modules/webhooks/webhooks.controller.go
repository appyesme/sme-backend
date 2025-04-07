package webhooks

import (
	"encoding/json"
	"io"
	"net/http"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/services/appointments_service"
	"sme-backend/src/services/webhook_service"
)

func RozarPayVerifyPayment(response http.ResponseWriter, request *http.Request) {
	webhook_body, err := io.ReadAll(request.Body)
	if err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Unable to read request body", err)
		return
	}

	webhook_signature := request.Header.Get("X-Razorpay-Signature")
	verified := webhook_service.VerifyRazorpayWebhookSignature(webhook_body, webhook_signature)

	if verified {
		var payload webhook_service.RazorpayWebhookPayload
		err = json.Unmarshal(webhook_body, &payload)
		if err != nil {
			helpers.HandleError(response, http.StatusBadRequest, "Malformed input", err)
			return
		}

		if payload.Payload.Payment.Entity.Status == "captured" {
			db := database.GetRlsContextDB(request)
			if err := appointments_service.UpdateAppointmentBookedPaymentStatus(db, payload); err != nil {
				helpers.HandleError(response, http.StatusBadRequest, "Malformed input", err)
				return
			}
		}
	}

	helpers.HandleSuccess(response, http.StatusOK, "ACK", nil)
}

func RazorpayRefundStatusListenerWebhook(response http.ResponseWriter, request *http.Request) {
	webhook_body, err := io.ReadAll(request.Body)
	if err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Unable to read request body", err)
		return
	}

	webhook_signature := request.Header.Get("X-Razorpay-Signature")
	verified := webhook_service.VerifyRazorpayWebhookSignature(webhook_body, webhook_signature)

	if verified {

		var rzpWebhookRefundReponse webhook_service.RzpWebhookRefundReponse
		err = json.Unmarshal(webhook_body, &rzpWebhookRefundReponse)
		if err != nil {
			helpers.HandleError(response, http.StatusBadRequest, "Malformed input", err)
			return
		}

		if rzpWebhookRefundReponse.Entity == "refund" {
			db := database.GetRlsContextDB(request)
			if err := appointments_service.UpdateAppointmentRefundPaymentStatus(db, rzpWebhookRefundReponse); err != nil {
				helpers.HandleError(response, http.StatusBadRequest, "Malformed input", err)
				return
			}
		}
	}

	helpers.HandleSuccess(response, http.StatusOK, "ACK", nil)
}
