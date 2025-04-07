package payment_status

// Define constants for StatusType with string values
const (
	PENDING          string = "PENDING"          // Appointment booking started but payment is pending.
	PAID             string = "PAID"             // Payment has been made by the customer to the entrepreneur account.
	REFUND_INITIATED string = "REFUND_INITIATED" // When entrepreneur rejected the appointment, the refund is initiated.
	REFUND_FAILED    string = "REFUND_FAILED"    //
	REFUND_SETTLED   string = "REFUND_SETTLED"   // When the refund is successful to the customer.
	SETTLED          string = "SETTLED"          // Payment has been settled to the entrepreneur account after taking the commission.
)
