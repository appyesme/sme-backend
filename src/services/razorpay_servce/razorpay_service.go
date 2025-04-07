package razorpay_servce

import (
	"errors"
	"fmt"
	"sme-backend/model"
	"sme-backend/src/core/config"
	"sme-backend/src/enums/razorpay_refund_status"

	"github.com/razorpay/razorpay-go"
)

func CreateRazorPayOrderForServicePrice(registration_fees float64, service_id, appointment_id, user_id string) (map[string]any, error) {
	razorpay_client := razorpay.NewClient(config.Config("RAZORPAY_KEY_ID"), config.Config("RAZORPAY_SECRET_KEY"))
	order := map[string]any{
		"amount":   registration_fees * 100, // registration_fees is in rupee. we need to convert that to paise.
		"currency": "INR",
		"receipt":  service_id,
		"notes": map[string]string{
			"service_id":     service_id,
			"created_by":     user_id,
			"appointment_id": appointment_id,
		},
	}

	return razorpay_client.Order.Create(order, nil)
}

func VerifyPaymentByOrderID(order_id string) (bool, error) {
	razorpay_client := razorpay.NewClient(config.Config("RAZORPAY_KEY_ID"), config.Config("RAZORPAY_SECRET_KEY"))
	// Fetch all payments related to the order_id
	order_details, err := razorpay_client.Order.Fetch(order_id, nil, nil)
	if err != nil {
		return false, errors.New("failed to verify payment")
	}

	return order_details["status"] == "paid", nil
}

type RefundResponse struct {
	PaymentStatus string `json:"payment_status"`
	RefundID      string `json:"refund_id"`
}

func InitiatedInstantPaymentRefund(payment model.Payment) (RefundResponse, error) {
	var refund_response RefundResponse

	razorpay_client := razorpay.NewClient(config.Config("RAZORPAY_KEY_ID"), config.Config("RAZORPAY_SECRET_KEY"))

	data := map[string]any{
		"speed":   "optimum",
		"receipt": fmt.Sprintf("Refund appoinment amount %v", payment.Amount),
		"notes": map[string]any{
			"appointment_id": payment.AppointmentID,
			"payment_id":     payment.PaymentID,
			"order_id":       payment.OrderID,
			"service_id":     payment.ServiceID,
			"created_by":     payment.CreatedBy,
		},
	}

	// RazorPay Refund Possible Statuses: "pending", "processed", "failed"
	refund_response_payload, err := razorpay_client.Payment.Refund(payment.PaymentID, int(payment.Amount), data, nil)
	if err != nil {
		return refund_response, err
	}

	payment_status, err := razorpay_refund_status.GetPaymentStatusBasedRazorpayRefundStatus(refund_response_payload["status"].(string))
	if err != nil {
		return refund_response, err
	}

	refund_response = RefundResponse{
		RefundID:      refund_response_payload["id"].(string),
		PaymentStatus: payment_status,
	}

	return refund_response, nil
}
