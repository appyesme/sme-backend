package users_handler

import (
	"encoding/json"

	"github.com/lib/pq"
)

type GetUserAppointmentsDto struct {
	ID                string          `json:"id"`
	CreatedBy         string          `json:"created_by"`
	UpdatedAt         string          `json:"updated_at"`
	ServiceID         string          `json:"service_id"`
	ServiceTimingID   string          `json:"service_timing_id"`
	AppointmentDate   string          `json:"appointment_date"`
	StartTime         string          `json:"start_time"`
	EndTime           string          `json:"end_time"`
	Status            string          `json:"status"`
	HomeServiceNeeded bool            `json:"home_service_needed"`
	Candidate         json.RawMessage `json:"candidate"`
	Service           json.RawMessage `json:"service"`
	Payment           json.RawMessage `json:"payment"`
}

type UserDetailsDto struct {
	ID                  string          `json:"id"`
	PhoneNumber         string          `json:"phone_number"`
	Name                string          `json:"name"`
	Email               string          `json:"email"`
	UserType            string          `json:"user_type"`
	PhotoURL            string          `json:"photo_url"`
	Expertises          json.RawMessage `json:"expertises"`
	TotalWorkExperience int16           `json:"total_work_experience"`
	About               string          `json:"about"`
	Favourited          bool            `json:"favourited"`
}

type UpdateUserDetailsDto struct {
	Name                *string         `json:"name"`
	Email               *string         `json:"email"`
	Expertises          *pq.StringArray `json:"expertises"`
	TotalWorkExperience *int16          `json:"total_work_experience"`
	About               *string         `json:"about"`
}
