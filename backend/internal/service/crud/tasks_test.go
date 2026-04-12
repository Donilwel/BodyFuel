package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── inline mocks ───────────────────────────────────────────────────────────

type mockTasksRepository struct{ mock.Mock }

func (m *mockTasksRepository) List(ctx context.Context, f dto.TasksFilter, withBlock bool) ([]*entities.Task, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *mockTasksRepository) Get(ctx context.Context, f dto.TasksFilter, withBlock bool) (*entities.Task, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Task), args.Error(1)
}

func (m *mockTasksRepository) Update(ctx context.Context, t *entities.Task) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTasksRepository) Delete(ctx context.Context, ids []uuid.UUID) error {
	return m.Called(ctx, ids).Error(0)
}

// passThroughTxManager is shared by all crud test files in this package.
type passThroughTxManager struct{}

func (p *passThroughTxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// ── helpers ────────────────────────────────────────────────────────────────

func newTaskService(tasksRepo *mockTasksRepository) *Service {
	return &Service{
		transactionManager: &passThroughTxManager{},
		tasksRepository:    tasksRepo,
	}
}

func newTestTask(id uuid.UUID, taskType entities.TaskType) *entities.Task {
	return entities.NewTask(entities.WithTaskRestoreSpec(entities.TaskRestoreSpecification{
		UUID:        id,
		TypeNm:      taskType,
		State:       entities.TaskStateRunning,
		MaxAttempts: 3,
		Attempts:    0,
		RetryAt:     time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}))
}

// ── GetTask ────────────────────────────────────────────────────────────────

func TestGetTask_Success(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	task := newTestTask(id, entities.TaskTypeSendCodeOnEmail)

	repo := &mockTasksRepository{}
	repo.On("Get", mock.Anything, dto.TasksFilter{IDs: []uuid.UUID{id}}, false).Return(task, nil)

	svc := newTaskService(repo)
	got, err := svc.GetTask(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, task, got)
	repo.AssertExpectations(t)
}

func TestGetTask_NotFound(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	repo := &mockTasksRepository{}
	repo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))

	svc := newTaskService(repo)
	_, err := svc.GetTask(ctx, id)
	assert.Error(t, err)
}

// ── ListTasks ──────────────────────────────────────────────────────────────

func TestListTasks_ReturnsList(t *testing.T) {
	ctx := context.Background()
	tasks := []*entities.Task{
		newTestTask(uuid.New(), entities.TaskTypeSendCodeOnEmail),
		newTestTask(uuid.New(), entities.TaskTypeSendCodeOnPhone),
	}

	repo := &mockTasksRepository{}
	repo.On("List", mock.Anything, dto.TasksFilter{}, false).Return(tasks, nil)

	svc := newTaskService(repo)
	got, err := svc.ListTasks(ctx, dto.TasksFilter{})

	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestListTasks_RepoError(t *testing.T) {
	ctx := context.Background()

	repo := &mockTasksRepository{}
	repo.On("List", mock.Anything, mock.Anything, false).Return(nil, errors.New("db error"))

	svc := newTaskService(repo)
	_, err := svc.ListTasks(ctx, dto.TasksFilter{})
	assert.Error(t, err)
}

// ── RestartTask ────────────────────────────────────────────────────────────

func TestRestartTask_Success(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	task := newTestTask(id, entities.TaskTypeSendNotificationEmail)
	task.Failed()

	repo := &mockTasksRepository{}
	repo.On("Get", mock.Anything, dto.TasksFilter{IDs: []uuid.UUID{id}}, true).Return(task, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(t *entities.Task) bool {
		return t.State() == entities.TaskStateRunning
	})).Return(nil)

	svc := newTaskService(repo)
	err := svc.RestartTask(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, entities.TaskStateRunning, task.State())
	repo.AssertExpectations(t)
}

func TestRestartTask_GetError(t *testing.T) {
	ctx := context.Background()

	repo := &mockTasksRepository{}
	repo.On("Get", mock.Anything, mock.Anything, true).Return(nil, errors.New("not found"))

	svc := newTaskService(repo)
	err := svc.RestartTask(ctx, uuid.New())
	assert.Error(t, err)
}

// ── DeleteTask ─────────────────────────────────────────────────────────────

func TestDeleteTask_Success(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	repo := &mockTasksRepository{}
	repo.On("Delete", mock.Anything, []uuid.UUID{id}).Return(nil)

	svc := newTaskService(repo)
	err := svc.DeleteTask(ctx, id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteTask_RepoError(t *testing.T) {
	ctx := context.Background()

	repo := &mockTasksRepository{}
	repo.On("Delete", mock.Anything, mock.Anything).Return(errors.New("db error"))

	svc := newTaskService(repo)
	err := svc.DeleteTask(ctx, uuid.New())
	assert.Error(t, err)
}
