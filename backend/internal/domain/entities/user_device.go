package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserDevice struct {
	id          uuid.UUID
	userID      uuid.UUID
	deviceToken string
	platform    string
	createdAt   time.Time
	updatedAt   time.Time
}

func (d *UserDevice) ID() uuid.UUID       { return d.id }
func (d *UserDevice) UserID() uuid.UUID   { return d.userID }
func (d *UserDevice) DeviceToken() string { return d.deviceToken }
func (d *UserDevice) Platform() string    { return d.platform }
func (d *UserDevice) CreatedAt() time.Time { return d.createdAt }
func (d *UserDevice) UpdatedAt() time.Time { return d.updatedAt }

type UserDeviceInitSpec struct {
	UserID      uuid.UUID
	DeviceToken string
	Platform    string
}

type UserDeviceRestoreSpec struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	DeviceToken string
	Platform    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewUserDevice(spec UserDeviceInitSpec) *UserDevice {
	return &UserDevice{
		id:          uuid.New(),
		userID:      spec.UserID,
		deviceToken: spec.DeviceToken,
		platform:    spec.Platform,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
}

func RestoreUserDevice(spec UserDeviceRestoreSpec) *UserDevice {
	return &UserDevice{
		id:          spec.ID,
		userID:      spec.UserID,
		deviceToken: spec.DeviceToken,
		platform:    spec.Platform,
		createdAt:   spec.CreatedAt,
		updatedAt:   spec.UpdatedAt,
	}
}
