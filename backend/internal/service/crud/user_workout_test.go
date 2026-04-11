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

func TestService_GetWorkout(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	workout := entities.NewWorkout(entities.WithWorkoutRestoreSpec(entities.WorkoutRestoreSpec{
		ID:        id,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}))

	tests := []struct {
		name      string
		filter    dto.WorkoutsFilter
		withBlock bool
		mock      func(repo *mocks.WorkoutsRepository)
		wantErr   bool
	}{
		{
			name:      "found",
			filter:    dto.WorkoutsFilter{ID: &id, UserID: &userID},
			withBlock: false,
			mock: func(repo *mocks.WorkoutsRepository) {
				repo.On("Get", mock.Anything, mock.Anything, false).Return(workout, nil)
			},
			wantErr: false,
		},
		{
			name:      "not found",
			filter:    dto.WorkoutsFilter{ID: &id},
			withBlock: false,
			mock: func(repo *mocks.WorkoutsRepository) {
				repo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.WorkoutsRepository{}
			tt.mock(repo)

			s := NewService(&Config{WorkoutsRepository: repo})

			got, err := s.GetWorkout(ctx, tt.filter, tt.withBlock)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, id, got.ID())
			}
		})
	}
}

func TestService_UpdateWorkoutByFilter(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	status := entities.WorkoutStatusDone

	workout := entities.NewWorkout(entities.WithWorkoutRestoreSpec(entities.WorkoutRestoreSpec{
		ID:        id,
		UserID:    userID,
		Status:    entities.WorkoutStatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}))

	tests := []struct {
		name    string
		filter  dto.WorkoutsFilter
		params  entities.WorkoutUpdateParams
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.WorkoutsRepository)
		wantErr bool
	}{
		{
			name:   "success",
			filter: dto.WorkoutsFilter{ID: &id, UserID: &userID},
			params: entities.WorkoutUpdateParams{Status: &status},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.WorkoutsRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything, true).Return(workout, nil)
						repo.On("Update", mock.Anything, mock.Anything).Return(nil)
						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name:   "workout not found",
			filter: dto.WorkoutsFilter{ID: &id},
			params: entities.WorkoutUpdateParams{Status: &status},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.WorkoutsRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything, true).Return(nil, errors.New("not found"))
						return fn(ctx)
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.WorkoutsRepository{}
			tx := &mocks.TransactionManager{}
			tt.mockTx(tx, repo)

			s := NewService(&Config{WorkoutsRepository: repo, TransactionManager: tx})

			err := s.UpdateWorkoutByFilter(ctx, tt.filter, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ListWorkouts(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	w1 := entities.NewWorkout(entities.WithWorkoutRestoreSpec(entities.WorkoutRestoreSpec{
		ID: uuid.New(), UserID: userID, CreatedAt: now, UpdatedAt: now,
	}))
	w2 := entities.NewWorkout(entities.WithWorkoutRestoreSpec(entities.WorkoutRestoreSpec{
		ID: uuid.New(), UserID: userID, CreatedAt: now, UpdatedAt: now,
	}))

	tests := []struct {
		name    string
		filter  dto.WorkoutsFilter
		mock    func(repo *mocks.WorkoutsRepository)
		wantLen int
		wantErr bool
	}{
		{
			name:   "returns workouts",
			filter: dto.WorkoutsFilter{UserID: &userID},
			mock: func(repo *mocks.WorkoutsRepository) {
				repo.On("TopListWithLimit", mock.Anything, mock.Anything, 0, false).
					Return([]*entities.Workout{w1, w2}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:   "repo error",
			filter: dto.WorkoutsFilter{UserID: &userID},
			mock: func(repo *mocks.WorkoutsRepository) {
				repo.On("TopListWithLimit", mock.Anything, mock.Anything, 0, false).
					Return(nil, errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.WorkoutsRepository{}
			tt.mock(repo)

			s := NewService(&Config{WorkoutsRepository: repo})

			got, err := s.ListWorkouts(ctx, tt.filter, false)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}
