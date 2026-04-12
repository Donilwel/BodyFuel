package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	internalerrors "backend/internal/errors"
	"backend/internal/service/crud/mocks"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── helpers ────────────────────────────────────────────────────────────────

func newParamsService(paramsRepo *mocks.UserParamsRepository) *Service {
	return &Service{
		transactionManager:   &passThroughTxManager{},
		userParamsRepository: paramsRepo,
	}
}

func newTestUserParams(userID uuid.UUID) *entities.UserParams {
	return entities.NewUserParams(entities.WithUserParamsInitSpec(entities.UserParamsInitSpec{
		ID:     uuid.New(),
		UserID: userID,
		Height: 175,
		Wants:  entities.LoseWeight,
	}))
}

// ── GetParamsUser ──────────────────────────────────────────────────────────

func TestGetParamsUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newTestUserParams(userID)
	f := dto.UserParamsFilter{UserID: &userID}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(params, nil)

	svc := newParamsService(repo)
	got, err := svc.GetParamsUser(ctx, f, false)

	assert.NoError(t, err)
	assert.Equal(t, params, got)
}

func TestGetParamsUser_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserParamsFilter{UserID: &userID}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(nil, errors.New("not found"))

	svc := newParamsService(repo)
	_, err := svc.GetParamsUser(ctx, f, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user params")
}

// ── CreateParamsUser ───────────────────────────────────────────────────────

func TestCreateParamsUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	spec := entities.UserParamsInitSpec{
		ID:     uuid.New(),
		UserID: userID,
		Height: 180,
		Wants:  entities.BuildMuscle,
	}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, dto.UserParamsFilter{UserID: &userID}, false).
		Return(nil, errors.New("not found"))
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserParams")).Return(nil)

	svc := newParamsService(repo)
	err := svc.CreateParamsUser(ctx, spec)

	assert.NoError(t, err)
}

func TestCreateParamsUser_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	spec := entities.UserParamsInitSpec{
		ID:     uuid.New(),
		UserID: userID,
	}

	existing := newTestUserParams(userID)

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, dto.UserParamsFilter{UserID: &userID}, false).
		Return(existing, nil)

	svc := newParamsService(repo)
	err := svc.CreateParamsUser(ctx, spec)

	assert.Error(t, err)
	assert.ErrorIs(t, err, internalerrors.ErrUserParamsAlreadyExists)
}

func TestCreateParamsUser_CreateRepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	spec := entities.UserParamsInitSpec{
		ID:     uuid.New(),
		UserID: userID,
	}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, dto.UserParamsFilter{UserID: &userID}, false).
		Return(nil, errors.New("not found"))
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	svc := newParamsService(repo)
	err := svc.CreateParamsUser(ctx, spec)

	assert.Error(t, err)
}

// ── UpdateParamsUser ───────────────────────────────────────────────────────

func TestUpdateParamsUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserParamsFilter{UserID: &userID}
	params := newTestUserParams(userID)

	newHeight := 190
	updateParams := entities.UserParamsUpdateParams{Height: &newHeight}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(params, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.UserParams")).Return(nil)

	svc := newParamsService(repo)
	err := svc.UpdateParamsUser(ctx, f, updateParams)

	assert.NoError(t, err)
}

func TestUpdateParamsUser_GetError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserParamsFilter{UserID: &userID}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(nil, errors.New("not found"))

	svc := newParamsService(repo)
	err := svc.UpdateParamsUser(ctx, f, entities.UserParamsUpdateParams{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user params")
}

func TestUpdateParamsUser_UpdateRepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserParamsFilter{UserID: &userID}
	params := newTestUserParams(userID)

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Get", mock.Anything, f, false).Return(params, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	svc := newParamsService(repo)
	err := svc.UpdateParamsUser(ctx, f, entities.UserParamsUpdateParams{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update")
}

// ── DeleteParamsUser ───────────────────────────────────────────────────────

func TestDeleteParamsUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserParamsFilter{UserID: &userID}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Delete", mock.Anything, f).Return(nil)

	svc := newParamsService(repo)
	err := svc.DeleteParamsUser(ctx, f)

	assert.NoError(t, err)
}

func TestDeleteParamsUser_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	f := dto.UserParamsFilter{UserID: &userID}

	repo := mocks.NewUserParamsRepository(t)
	repo.On("Delete", mock.Anything, f).Return(errors.New("db error"))

	svc := newParamsService(repo)
	err := svc.DeleteParamsUser(ctx, f)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete user params")
}
