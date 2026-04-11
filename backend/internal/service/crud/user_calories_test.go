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

func TestService_CreateUserCalories(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name    string
		spec    entities.UserCaloriesInitSpec
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository)
		wantErr bool
	}{
		{
			name: "success",
			spec: entities.UserCaloriesInitSpec{
				ID:       uuid.New(),
				UserID:   userID,
				Calories: 500,
				Date:     now,
			},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Create", mock.Anything, mock.Anything).Return(nil)
						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name: "repo error",
			spec: entities.UserCaloriesInitSpec{
				ID:       uuid.New(),
				UserID:   userID,
				Calories: 200,
				Date:     now,
			},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
						return fn(ctx)
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.UserCaloriesRepository{}
			tx := &mocks.TransactionManager{}
			tt.mockTx(tx, repo)

			s := NewService(&Config{
				UserCaloriesRepository: repo,
				TransactionManager:     tx,
			})

			err := s.CreateUserCalories(ctx, tt.spec)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ListUserCalories(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	entry := entities.NewUserCalories(entities.WithUserCaloriesRestoreSpec(entities.UserCaloriesRestoreSpec{
		ID:        uuid.New(),
		UserID:    userID,
		Calories:  300,
		Date:      now,
		CreatedAt: now,
		UpdatedAt: now,
	}))

	tests := []struct {
		name    string
		filter  dto.UserCaloriesFilter
		mock    func(repo *mocks.UserCaloriesRepository)
		wantLen int
		wantErr bool
	}{
		{
			name:   "returns list",
			filter: dto.UserCaloriesFilter{UserID: &userID},
			mock: func(repo *mocks.UserCaloriesRepository) {
				repo.On("List", mock.Anything, mock.Anything).
					Return([]*entities.UserCalories{entry}, nil)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:   "empty list",
			filter: dto.UserCaloriesFilter{UserID: &userID},
			mock: func(repo *mocks.UserCaloriesRepository) {
				repo.On("List", mock.Anything, mock.Anything).
					Return([]*entities.UserCalories{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:   "repo error",
			filter: dto.UserCaloriesFilter{UserID: &userID},
			mock: func(repo *mocks.UserCaloriesRepository) {
				repo.On("List", mock.Anything, mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.UserCaloriesRepository{}
			tt.mock(repo)

			s := NewService(&Config{UserCaloriesRepository: repo})

			got, err := s.ListUserCalories(ctx, tt.filter)
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

func TestService_UpdateUserCalories(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	entryID := uuid.New()
	now := time.Now()
	newCals := 600

	existing := entities.NewUserCalories(entities.WithUserCaloriesRestoreSpec(entities.UserCaloriesRestoreSpec{
		ID:        entryID,
		UserID:    userID,
		Calories:  300,
		Date:      now,
		CreatedAt: now,
		UpdatedAt: now,
	}))

	tests := []struct {
		name    string
		filter  dto.UserCaloriesFilter
		params  entities.UserCaloriesUpdateParams
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository)
		wantErr bool
	}{
		{
			name:   "success",
			filter: dto.UserCaloriesFilter{ID: &entryID, UserID: &userID},
			params: entities.UserCaloriesUpdateParams{Calories: &newCals},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything).Return(existing, nil)
						repo.On("Update", mock.Anything, mock.Anything).Return(nil)
						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name:   "not found",
			filter: dto.UserCaloriesFilter{ID: &entryID, UserID: &userID},
			params: entities.UserCaloriesUpdateParams{Calories: &newCals},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
						return fn(ctx)
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.UserCaloriesRepository{}
			tx := &mocks.TransactionManager{}
			tt.mockTx(tx, repo)

			s := NewService(&Config{
				UserCaloriesRepository: repo,
				TransactionManager:     tx,
			})

			err := s.UpdateUserCalories(ctx, tt.filter, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_DeleteUserCalories(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	entryID := uuid.New()

	tests := []struct {
		name    string
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository)
		wantErr bool
	}{
		{
			name: "success",
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						repo.On("Delete", mock.Anything, mock.Anything).Return(nil)
						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name: "repo error",
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserCaloriesRepository) {
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
			repo := &mocks.UserCaloriesRepository{}
			tx := &mocks.TransactionManager{}
			tt.mockTx(tx, repo)

			s := NewService(&Config{
				UserCaloriesRepository: repo,
				TransactionManager:     tx,
			})

			err := s.DeleteUserCalories(ctx, entryID, userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
