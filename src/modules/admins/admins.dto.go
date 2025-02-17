package admins_handler

import (
	"encoding/json"

	"github.com/lib/pq"
)

type PaymentClearanceEntrepreneursDTO struct {
	ID            string          `json:"id"`
	PhoneNumber   string          `json:"phone_number"`
	Name          string          `json:"name"`
	LastClearedAt *string         `gorm:"last_cleared_at" json:"last_cleared_at"`
	BankAccount   json.RawMessage `gorm:"bank_account" json:"bank_account"`
}

type UserDetails struct {
	ID                  string         `json:"id"`
	PhoneNumber         string         `json:"phone_number"`
	UserType            string         `json:"user_type"`
	VerifiedAt          *string        `json:"verified_at"`
	Name                string         `json:"name"`
	Email               string         `json:"email"`
	TotalWorkExperience int16          `json:"total_work_experience"`
	Expertises          pq.StringArray `gorm:"type:text[]" json:"expertises"`
	Documents           pq.StringArray `gorm:"type:text[]" json:"documents"`
	Verified            bool           `json:"verified"`
	AadharNumber        string         `json:"aadhar_number"`
	PanNumber           string         `json:"pan_number"`
}
