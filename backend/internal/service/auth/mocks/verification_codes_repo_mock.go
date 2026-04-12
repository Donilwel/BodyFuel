package mocks

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"

	"github.com/stretchr/testify/mock"
)

type UserVerificationCodesRepository struct {
	mock.Mock
}

func (_m *UserVerificationCodesRepository) Create(ctx context.Context, c *entities.UserVerificationCode) error {
	ret := _m.Called(ctx, c)
	return ret.Error(0)
}

func (_m *UserVerificationCodesRepository) GetLatest(ctx context.Context, f dto.UserVerificationCodeFilter) (*entities.UserVerificationCode, error) {
	ret := _m.Called(ctx, f)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*entities.UserVerificationCode), ret.Error(1)
}

func (_m *UserVerificationCodesRepository) MarkUsed(ctx context.Context, id interface{}) error {
	ret := _m.Called(ctx, id)
	return ret.Error(0)
}

func NewUserVerificationCodesRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserVerificationCodesRepository {
	m := &UserVerificationCodesRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
