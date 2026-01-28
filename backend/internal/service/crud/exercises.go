package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/errors"
	"context"
	"fmt"
	"github.com/google/uuid"
)

func (s *Service) GetExercise(ctx context.Context, f dto.ExerciseFilter, withBlock bool) (*entities.Exercise, error) {
	user, err := s.exercisesRepository.Get(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("get exercise: %w", err)
	}
	return user, nil
}

func (s *Service) CreateExercise(ctx context.Context, params entities.ExerciseInitSpec) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if _, err := s.exercisesRepository.Get(ctx, dto.ExerciseFilter{Name: &params.Name}, false); err == nil {
			return fmt.Errorf("create exercise: %w", errors.ErrExerciseAlreadyExists)
		}

		if err := s.exercisesRepository.Create(ctx, entities.NewExercise(entities.WithExerciseInitSpec(params))); err != nil {
			return fmt.Errorf("create exercise: %w", err)
		}
		return nil
	})
}

func (s *Service) UpdateExercise(ctx context.Context, f dto.ExerciseFilter, exercise entities.ExerciseUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		e, err := s.exercisesRepository.Get(ctx, f, false)
		if err != nil {
			return fmt.Errorf("update exercise: get exercise: %w", err)
		}
		e.Update(exercise)

		if err := s.exercisesRepository.Update(ctx, e); err != nil {
			return fmt.Errorf("update exercise: update: %w", err)
		}
		return nil
	})
}

func (s *Service) DeleteExercise(ctx context.Context, f dto.ExerciseFilter) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.exercisesRepository.Delete(ctx, f); err != nil {
			return fmt.Errorf("delete exercise: delete:%w", err)
		}
		return nil
	})
}

func (s *Service) ListExercise(ctx context.Context, userID uuid.UUID, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error) {
	user, err := s.userParamsRepository.Get(ctx, dto.UserParamsFilter{UserID: &userID}, false)
	if err != nil {
		fmt.Errorf("list exercise: get: %w", err)
	}

	userLevel, err := user.Lifestyle().ToLevelPreparation()
	if err != nil {
		fmt.Errorf("list exercise: to level preparation: %w", err)
	}
	f.LevelPreparation = &userLevel
	e, err := s.exercisesRepository.List(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("list exercise: %w", err)
	}

	return e, nil
}
