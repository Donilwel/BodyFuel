package dto

import "github.com/google/uuid"

type UserDeviceFilter struct {
	ID          *uuid.UUID
	UserID      *uuid.UUID
	DeviceToken *string
}
