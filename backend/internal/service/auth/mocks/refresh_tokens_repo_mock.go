package mocks

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type UserRefreshTokensRepository struct {
	mock.Mock
}

func (_m *UserRefreshTokensRepository) Create(ctx context.Context, t *entities.UserRefreshToken) error {
	ret := _m.Called(ctx, t)
	return ret.Error(0)
}

func (_m *UserRefreshTokensRepository) Get(ctx context.Context, f dto.UserRefreshTokenFilter) (*entities.UserRefreshToken, error) {
	ret := _m.Called(ctx, f)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*entities.UserRefreshToken), ret.Error(1)
}

func (_m *UserRefreshTokensRepository) Delete(ctx context.Context, f dto.UserRefreshTokenFilter) error {
	ret := _m.Called(ctx, f)
	return ret.Error(0)
}

func (_m *UserRefreshTokensRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)
	return ret.Error(0)
}

func NewUserRefreshTokensRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserRefreshTokensRepository {
	m := &UserRefreshTokensRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
