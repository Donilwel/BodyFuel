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

func TestService_GetInfoUser(t *testing.T) {
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()

	tests := []struct {
		name      string
		filter    dto.UserInfoFilter
		withBlock bool
		mockGet   func(repo *mocks.UserInfoRepository)
		wantErr   error
	}{
		{
			name:      "user exists",
			filter:    dto.UserInfoFilter{ID: &id1},
			withBlock: false,
			mockGet: func(repo *mocks.UserInfoRepository) {
				repo.On("Get", mock.Anything, mock.Anything, false).
					Return(entities.NewUserInfo(entities.WithUserInfoInitSpec(
						entities.UserInfoInitSpec{ID: id1})), nil)
			},
			wantErr: nil,
		},
		{
			name:      "repo returns error",
			filter:    dto.UserInfoFilter{ID: &id2},
			withBlock: true,
			mockGet: func(repo *mocks.UserInfoRepository) {
				repo.On("Get", mock.Anything, mock.Anything, true).
					Return(nil, errors.New("db error"))
			},
			wantErr: errors.New("get user info: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.UserInfoRepository{}
			tt.mockGet(repo)

			s := NewService(&Config{
				UserInfoRepository: repo,
			})

			got, err := s.GetInfoUser(ctx, tt.filter, tt.withBlock)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *tt.filter.ID, got.ID())
			}
		})
	}
}

func TestService_CreateInfoUser(t *testing.T) {
	ctx := context.Background()
	id1 := uuid.New()

	tests := []struct {
		name    string
		info    entities.UserInfoInitSpec
		mockTx  func(tx *mocks.TransactionManager, repo *mocks.UserInfoRepository)
		wantErr error
	}{
		{
			name: "create success",
			info: entities.UserInfoInitSpec{ID: id1},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserInfoRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						// Возвращаем результат fn(ctx) для корректного flow
						repo.On("Get", mock.Anything, mock.Anything, false).
							Return(nil, errors.New("not found"))
						repo.On("Create", mock.Anything, mock.Anything).
							Return(nil)
						return fn(ctx)
					})
			},
			wantErr: nil,
		},
		{
			name: "user already exists",
			info: entities.UserInfoInitSpec{ID: id1},
			mockTx: func(tx *mocks.TransactionManager, repo *mocks.UserInfoRepository) {
				tx.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						repo.On("Get", mock.Anything, mock.Anything, false).
							Return(entities.NewUserInfo(entities.WithUserInfoInitSpec(
								entities.UserInfoInitSpec{ID: id1})), nil)
						// Возвращаем результат fn(ctx), чтобы сервис получил ErrUserInfoAlreadyExists
						return fn(ctx)
					})
			},
			wantErr: errs.ErrUserInfoAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.UserInfoRepository{}
			tx := &mocks.TransactionManager{}
			if tt.mockTx != nil {
				tt.mockTx(tx, repo)
			}

			s := NewService(&Config{
				UserInfoRepository: repo,
				TransactionManager: tx,
			})

			err := s.CreateInfoUser(ctx, tt.info)

			if tt.wantErr != nil {
				// assert.ErrorIs теперь корректно отрабатывает
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
