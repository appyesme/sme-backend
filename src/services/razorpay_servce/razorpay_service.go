package razorpay_servce

import (
	"errors"
	"sme-backend/src/core/config"

	"github.com/razorpay/razorpay-go"
)

func CreateRazorPayOrderForServicePrice(
	registration_fees float64,
	service_id string,
	appointment_id string,
	user_id string,
) (map[string]any, error) {
	razorpay_client := razorpay.NewClient(config.Config("RAZORPAY_KEY_ID"), config.Config("RAZORPAY_SECRET_KEY"))
	order := map[string]any{
		"amount":   registration_fees * 100,
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
