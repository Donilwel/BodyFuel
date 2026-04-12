package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/service/crud/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── helpers ────────────────────────────────────────────────────────────────

func newWeightService(weightRepo *mocks.UserWeightRepository) *Service {
	return &Service{
		transactionManager:  &passThroughTxManager{},
		userWeightRepository: weightRepo,
	}
}

func newTestUserWeight(userID uuid.UUID) *entities.UserWeight {
	return entities.NewUserWeight(entities.WithUserWeightRestoreSpec(entities.UserWeightRestoreSpec{
		ID:     uuid.New(),
		UserID: userID,
		Weight: 75.5,
		Date:   time.Now(),
	}))
}

// ── GetWeightUser ──────────────────────────────────────────────────────────

func TestGetWeightUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	uw := newTestUserWeight(userID)
	f := dto.UserWeightFilter{UserID: &userID}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(uw, nil)

	svc := newWeightService(repo)
	got, err := svc.GetWeightUser(ctx, f, false)

	assert.NoError(t, err)
	assert.Equal(t, uw, got)
}

func TestGetWeightUser_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(nil, errors.New("not found"))

	svc := newWeightService(repo)
	_, err := svc.GetWeightUser(ctx, f, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user weight")
}

// ── ListWeightsUser ────────────────────────────────────────────────────────

func TestListWeightsUser_ReturnsList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}

	weights := []*entities.UserWeight{
		newTestUserWeight(userID),
		newTestUserWeight(userID),
	}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("List", mock.Anything, f, false).Return(weights, nil)

	svc := newWeightService(repo)
	got, err := svc.ListWeightsUser(ctx, f, false)

	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestListWeightsUser_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("List", mock.Anything, f, false).Return(nil, errors.New("db error"))

	svc := newWeightService(repo)
	_, err := svc.ListWeightsUser(ctx, f, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list user weight")
}

// ── CreateWeightUser ───────────────────────────────────────────────────────

func TestCreateWeightUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	spec := entities.UserWeightInitSpec{
		ID:     uuid.New(),
		UserID: userID,
		Weight: 80.0,
		Date:   time.Now(),
	}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserWeight")).Return(nil)

	svc := newWeightService(repo)
	err := svc.CreateWeightUser(ctx, spec)

	assert.NoError(t, err)
}

func TestCreateWeightUser_RepoError(t *testing.T) {
	ctx := context.Background()
	spec := entities.UserWeightInitSpec{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Weight: 80.0,
	}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	svc := newWeightService(repo)
	err := svc.CreateWeightUser(ctx, spec)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create user weight")
}

// ── UpdateWeightUser ───────────────────────────────────────────────────────

func TestUpdateWeightUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}
	uw := newTestUserWeight(userID)

	newWeight := 78.0
	updateParams := entities.UserWeightUpdateParams{Weight: &newWeight}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(uw, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.UserWeight")).Return(nil)

	svc := newWeightService(repo)
	err := svc.UpdateWeightUser(ctx, f, updateParams)

	assert.NoError(t, err)
}

func TestUpdateWeightUser_GetError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(nil, errors.New("not found"))

	svc := newWeightService(repo)
	err := svc.UpdateWeightUser(ctx, f, entities.UserWeightUpdateParams{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user weight")
}

func TestUpdateWeightUser_UpdateRepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}
	uw := newTestUserWeight(userID)

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(uw, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	svc := newWeightService(repo)
	err := svc.UpdateWeightUser(ctx, f, entities.UserWeightUpdateParams{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update")
}

// ── DeleteWeightUser ───────────────────────────────────────────────────────

func TestDeleteWeightUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Delete", mock.Anything, f).Return(nil)

	svc := newWeightService(repo)
	err := svc.DeleteWeightUser(ctx, f)

	assert.NoError(t, err)
}

func TestDeleteWeightUser_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserWeightFilter{UserID: &userID}

	repo := mocks.NewUserWeightRepository(t)
	repo.On("Delete", mock.Anything, f).Return(errors.New("db error"))

	svc := newWeightService(repo)
	err := svc.DeleteWeightUser(ctx, f)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete user weight")
}
