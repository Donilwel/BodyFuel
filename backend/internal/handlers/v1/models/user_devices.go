package models

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type RegisterDeviceRequest struct {
	DeviceToken string `json:"device_token" validate:"required"`
	Platform    string `json:"platform" validate:"required,oneof=ios android"`
}

type UserDeviceResponse struct {
	ID          uuid.UUID `json:"id"`
	DeviceToken string    `json:"device_token"`
	Platform    string    `json:"platform"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewUserDeviceResponse(d *entities.UserDevice) UserDeviceResponse {
	return UserDeviceResponse{
		ID:          d.ID(),
		DeviceToken: d.DeviceToken(),
		Platform:    d.Platform(),
		CreatedAt:   d.CreatedAt(),
	}
}

func NewUserDevicesResponse(devices []*entities.UserDevice) []UserDeviceResponse {
	result := make([]UserDeviceResponse, len(devices))
	for i, d := range devices {
		result[i] = NewUserDeviceResponse(d)
	}
	return result
}
