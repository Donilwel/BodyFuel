package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	errs "backend/internal/errors"
	"backend/internal/service/crud/mocks"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_GetExercise(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	exercise := entities.NewExercise(entities.WithExerciseInitSpec(entities.ExerciseInitSpec{
		ID:   id,
		Name: "Push-up",
	}))

	tests := []struct {
		name      string
		filter    dto.ExerciseFilter
		withBlock bool
		mock      func(repo *mocks.ExercisesRepository)
		wantErr   bool
	}{
		{
			name:      "found",
			filter:    dto.ExerciseFilter{ID: &id},
			withBlock: false,
			mock: func(repo *mocks.ExercisesRepository) {
				repo.On("Get", mock.Anything, mock.Anything, false).Return(exercise, nil)
			},
			wantErr: false,
		},
		{
			name:      "not found",
			filter:    dto.ExerciseFilter{ID: &id},
			withBlock: false,
			mock: func(repo *mocks.ExercisesRepository) {
				repo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.ExercisesRepository{}
			tt.mock(repo)

			s := NewService(&Config{ExercisesRepository: repo})

			got, err := s.GetExercise(ctx, tt.filter, tt.withBlock)
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

func TestService_CreateExercise(t *testing.T) {
	ctx := context.Background()

	spec := entities.ExerciseInitSpec{
		ID:   uuid.New(),
		Name: "Squat",
	}

	tests := []struct {
		name    string
		spec    entities.ExerciseInitSpec
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository)
		wantErr error
	}{
		{
			name: "success",
			spec: spec,
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))
						repo.On("Create", mock.Anything, mock.Anything).Return(nil)
						return fn(ctx)
					})
			},
			wantErr: nil,
		},
		{
			name: "already exists",
			spec: spec,
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						existing := entities.NewExercise(entities.WithExerciseInitSpec(spec))
						repo.On("Get", mock.Anything, mock.Anything, false).Return(existing, nil)
						return fn(ctx)
					})
			},
			wantErr: errs.ErrExerciseAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.ExercisesRepository{}
			tx := &mocks.TransactionManager{}
			tt.mockTx(tx, repo)

			s := NewService(&Config{ExercisesRepository: repo, TransactionManager: tx})

			err := s.CreateExercise(ctx, tt.spec)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_UpdateExercise(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	newName := "Pull-up"

	existing := entities.NewExercise(entities.WithExerciseInitSpec(entities.ExerciseInitSpec{
		ID:   id,
		Name: "Push-up",
	}))

	tests := []struct {
		name    string
		filter  dto.ExerciseFilter
		params  entities.ExerciseUpdateParams
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository)
		wantErr bool
	}{
		{
			name:   "success",
			filter: dto.ExerciseFilter{ID: &id},
			params: entities.ExerciseUpdateParams{Name: &newName},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything, false).Return(existing, nil)
						repo.On("Update", mock.Anything, mock.Anything).Return(nil)
						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name:   "not found",
			filter: dto.ExerciseFilter{ID: &id},
			params: entities.ExerciseUpdateParams{Name: &newName},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))
						return fn(ctx)
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.ExercisesRepository{}
			tx := &mocks.TransactionManager{}
			tt.mockTx(tx, repo)

			s := NewService(&Config{ExercisesRepository: repo, TransactionManager: tx})

			err := s.UpdateExercise(ctx, tt.filter, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_DeleteExercise(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	tests := []struct {
		name    string
		filter  dto.ExerciseFilter
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository)
		wantErr bool
	}{
		{
			name:   "success",
			filter: dto.ExerciseFilter{ID: &id},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Delete", mock.Anything, mock.Anything).Return(nil)
						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name:   "repo error",
			filter: dto.ExerciseFilter{ID: &id},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.ExercisesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Delete", mock.Anything, mock.Anything).Return(errors.New("db error"))
						return fn(ctx)
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.ExercisesRepository{}
			tx := &mocks.TransactionManager{}
			tt.mockTx(tx, repo)

			s := NewService(&Config{ExercisesRepository: repo, TransactionManager: tx})

			err := s.DeleteExercise(ctx, tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
