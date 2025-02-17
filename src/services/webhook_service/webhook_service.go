package webhook_service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sme-backend/src/core/config"
)

type RazorpayWebhookPayload struct {
	Payload struct {
		Payment struct {
			Entity struct {
				ID      string `json:"id"`
				Status  string `json:"status"`
				OrderID string `json:"order_id"`
				Receipt string `json:"receipt"`
				Notes   struct {
					AppointmentID string `json:"appointment_id"`
					ServiceID     string `json:"service_id"`
					CreatedBy     string `json:"created_by"`
				}
			}
		}
	}
}

func VerifyRazorpayWebhookSignature(webhook_body []byte, webhook_signature string) bool {
	// Create HMAC-SHA256 hash using the Razorpay Webhook Secret
	h := hmac.New(sha256.New, []byte(config.Config("RAZORPAY_WEBHOOK_SECRET")))
	h.Write([]byte(webhook_body))
	generated_signature := hex.EncodeToString(h.Sum(nil))
	// Compare generated signature with the provided signature
	return hmac.Equal([]byte(generated_signature), []byte(webhook_signature))
}
