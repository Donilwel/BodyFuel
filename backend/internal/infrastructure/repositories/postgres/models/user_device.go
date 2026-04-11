package models

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type UserDeviceRow struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	DeviceToken string    `db:"device_token"`
	Platform    string    `db:"platform"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewUserDeviceRow(d *entities.UserDevice) *UserDeviceRow {
	return &UserDeviceRow{
		ID:          d.ID(),
		UserID:      d.UserID(),
		DeviceToken: d.DeviceToken(),
		Platform:    d.Platform(),
		CreatedAt:   d.CreatedAt(),
		UpdatedAt:   d.UpdatedAt(),
	}
}

func (r *UserDeviceRow) ToEntity() *entities.UserDevice {
	return entities.RestoreUserDevice(entities.UserDeviceRestoreSpec{
		ID:          r.ID,
		UserID:      r.UserID,
		DeviceToken: r.DeviceToken,
		Platform:    r.Platform,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	})
}
