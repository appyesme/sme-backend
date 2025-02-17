package appointments_handler

import (
	"sme-backend/model"
)

type GetAppointmentsEnabledDayTimingsDto struct {
	model.ServiceTiming
	RemainingSlots string `json:"remaining_slots"`
}

type AppointmentBookDto struct {
	ServiceID         string `json:"service_id"`
	ServiceTimingID   string `json:"service_timing_id"`
	AppointmentDate   string `json:"appointment_date"`
	StartTime         string `json:"start_time"`
	EndTime           string `json:"end_time"`
	HomeServiceNeeded bool   `json:"home_service_needed"`
}

type AppointmentInitiatedDto struct {
	ServiceID     string  `json:"service_id"`
	AppointmentID string  `json:"appointment_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	OrderID       string  `json:"order_id"`
	Status        string  `json:"status"`
}

type GetAppointmentFinalPriceDto struct {
	HomeServiceNeeded bool `json:"home_service_needed"`
}
