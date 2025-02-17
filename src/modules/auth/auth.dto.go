package auth_handler

import (
	"github.com/lib/pq"
)

type VerifyOtpDto struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Otp         string `json:"otp" validate:"required"`
}

type SignInDto struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

type SignUpDto struct {
	PhoneNumber         string          `json:"phone_number" validate:"required"`
	UserType            string          `json:"user_type" validate:"required"`
	Name                string          `json:"name" validate:"required"`
	Email               *string         `json:"email"`
	TotalWorkExperience *int16          `json:"total_work_experience"`
	Expertises          *pq.StringArray `json:"expertises"`
	Documents           *pq.StringArray `json:"documents"`
	AadharNumber        *string         `json:"aadhar_number"`
	PanNumber           *string         `json:"pan_number"`
}
