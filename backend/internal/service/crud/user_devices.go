package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) RegisterUserDevice(ctx context.Context, spec entities.UserDeviceInitSpec) error {
	device := entities.NewUserDevice(spec)
	if err := s.userDevicesRepository.Upsert(ctx, device); err != nil {
		return fmt.Errorf("register user device: %w", err)
	}
	return nil
}

func (s *Service) ListUserDevices(ctx context.Context, userID uuid.UUID) ([]*entities.UserDevice, error) {
	devices, err := s.userDevicesRepository.List(ctx, dto.UserDeviceFilter{UserID: &userID})
	if err != nil {
		return nil, fmt.Errorf("list user devices: %w", err)
	}
	return devices, nil
}

func (s *Service) DeleteUserDevice(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.userDevicesRepository.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete user device: %w", err)
	}
	return nil
}
