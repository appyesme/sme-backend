package services_handler

import (
	"encoding/json"
	"sme-backend/model"
)

type GetServicesDto struct {
	model.Service
	Author json.RawMessage `json:"author"`
	Medias json.RawMessage `json:"medias"`
}

type GetBookingSlotDto struct {
	model.ServiceDay
	Timings json.RawMessage `json:"timings"`
}
