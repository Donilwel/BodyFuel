package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── inline mock for UserDevicesRepository ────────────────────────────────────

type mockUserDevicesRepository struct{ mock.Mock }

func (m *mockUserDevicesRepository) Upsert(ctx context.Context, device *entities.UserDevice) error {
	return m.Called(ctx, device).Error(0)
}

func (m *mockUserDevicesRepository) List(ctx context.Context, f dto.UserDeviceFilter) ([]*entities.UserDevice, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserDevice), args.Error(1)
}

func (m *mockUserDevicesRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

// ── helpers ────────────────────────────────────────────────────────────────

func newDeviceService(devicesRepo *mockUserDevicesRepository) *Service {
	return &Service{
		transactionManager:    &passThroughTxManager{},
		userDevicesRepository: devicesRepo,
	}
}

// ── RegisterUserDevice ─────────────────────────────────────────────────────

func TestRegisterUserDevice_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	spec := entities.UserDeviceInitSpec{
		UserID:      userID,
		DeviceToken: "token-abc",
		Platform:    "ios",
	}

	repo := &mockUserDevicesRepository{}
	repo.On("Upsert", mock.Anything, mock.AnythingOfType("*entities.UserDevice")).Return(nil)

	svc := newDeviceService(repo)
	err := svc.RegisterUserDevice(ctx, spec)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRegisterUserDevice_RepoError(t *testing.T) {
	ctx := context.Background()
	spec := entities.UserDeviceInitSpec{
		UserID:      uuid.New(),
		DeviceToken: "token-abc",
		Platform:    "android",
	}

	repo := &mockUserDevicesRepository{}
	repo.On("Upsert", mock.Anything, mock.Anything).Return(errors.New("db error"))

	svc := newDeviceService(repo)
	err := svc.RegisterUserDevice(ctx, spec)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "register user device")
}

// ── ListUserDevices ────────────────────────────────────────────────────────

func TestListUserDevices_ReturnsList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	device := entities.RestoreUserDevice(entities.UserDeviceRestoreSpec{
		ID:          uuid.New(),
		UserID:      userID,
		DeviceToken: "tok",
		Platform:    "ios",
	})

	repo := &mockUserDevicesRepository{}
	repo.On("List", mock.Anything, dto.UserDeviceFilter{UserID: &userID}).
		Return([]*entities.UserDevice{device}, nil)

	svc := newDeviceService(repo)
	got, err := svc.ListUserDevices(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, got, 1)
	repo.AssertExpectations(t)
}

func TestListUserDevices_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	repo := &mockUserDevicesRepository{}
	repo.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	svc := newDeviceService(repo)
	_, err := svc.ListUserDevices(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list user devices")
}

// ── DeleteUserDevice ───────────────────────────────────────────────────────

func TestDeleteUserDevice_Success(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	userID := uuid.New()

	repo := &mockUserDevicesRepository{}
	repo.On("Delete", mock.Anything, id, userID).Return(nil)

	svc := newDeviceService(repo)
	err := svc.DeleteUserDevice(ctx, id, userID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteUserDevice_RepoError(t *testing.T) {
	ctx := context.Background()

	repo := &mockUserDevicesRepository{}
	repo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))

	svc := newDeviceService(repo)
	err := svc.DeleteUserDevice(ctx, uuid.New(), uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete user device")
}
