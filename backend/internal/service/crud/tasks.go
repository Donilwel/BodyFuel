package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"fmt"
	"github.com/google/uuid"
)

func (s *Service) ListTasks(ctx context.Context, filter dto.TasksFilter) ([]*entities.Task, error) {
	ts, err := s.tasksRepository.List(ctx, filter, false)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	return ts, nil
}

func (s *Service) RestartTask(ctx context.Context, id uuid.UUID) error {
	if err := s.transactionManager.Do(ctx, func(ctx context.Context) error {
		t, err := s.tasksRepository.Get(ctx, dto.TasksFilter{
			IDs: []uuid.UUID{id},
		}, true)
		if err != nil {
			return fmt.Errorf("get task: %w", err)
		}

		t.Restart()

		if err = s.tasksRepository.Update(ctx, t); err != nil {
			return fmt.Errorf("update task: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	return nil
}

func (s *Service) DeleteTask(ctx context.Context, id uuid.UUID) error {
	if err := s.tasksRepository.Delete(ctx, []uuid.UUID{id}); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}
