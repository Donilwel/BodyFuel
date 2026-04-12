package mocks

import (
	"backend/internal/domain/entities"
	"context"

	"github.com/stretchr/testify/mock"
)

type TasksRepository struct {
	mock.Mock
}

func (_m *TasksRepository) Create(ctx context.Context, t *entities.Task) error {
	ret := _m.Called(ctx, t)
	return ret.Error(0)
}

func NewTasksRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *TasksRepository {
	m := &TasksRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
