package razorpay_refund_status

import (
	"errors"
	"sme-backend/src/enums/payment_status"
)

// RazorPay Refund Possible Statuses: "pending", "processed", "failed"
const (
	PENDING   string = "pending"
	PROCESSED string = "processed"
	FAILED    string = "failed"
)

// Get status from string
func GetPaymentStatusBasedRazorpayRefundStatus(status string) (string, error) {
	switch status {
	case PENDING:
		return payment_status.REFUND_INITIATED, nil
	case PROCESSED:
		return payment_status.REFUND_SETTLED, nil
	case FAILED:
		return payment_status.REFUND_FAILED, nil
	default:
		return "", errors.New("invalid status")
	}
}
