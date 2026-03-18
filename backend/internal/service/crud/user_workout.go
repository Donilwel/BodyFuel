package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"fmt"

	"github.com/google/uuid"
)

// WorkoutService реализует методы для работы с тренировками
func (s *Service) GetWorkout(ctx context.Context, f dto.WorkoutsFilter, withBlock bool) (*entities.Workout, error) {
	workout, err := s.workoutsRepository.Get(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("get workout: %w", err)
	}
	return workout, nil
}

func (s *Service) ListWorkouts(ctx context.Context, f dto.WorkoutsFilter, withBlock bool) ([]*entities.Workout, error) {
	workouts, err := s.workoutsRepository.TopListWithLimit(ctx, f, 0, withBlock)
	if err != nil {
		return nil, fmt.Errorf("list workouts: %w", err)
	}
	return workouts, nil
}

func (s *Service) CreateWorkout(ctx context.Context, workout *entities.Workout) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.workoutsRepository.Create(ctx, workout); err != nil {
			return fmt.Errorf("create workout: %w", err)
		}
		return nil
	})
}

func (s *Service) CreateWorkoutWithExercises(ctx context.Context, workout *entities.Workout, exercises []*entities.WorkoutsExercise) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.workoutsRepository.Create(ctx, workout); err != nil {
			return fmt.Errorf("create workout: %w", err)
		}

		for _, exercise := range exercises {
			if err := s.workoutsExerciseRepository.Create(ctx, exercise); err != nil {
				return fmt.Errorf("create workout exercise: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) UpdateWorkout(ctx context.Context, workout *entities.Workout) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.workoutsRepository.Update(ctx, workout); err != nil {
			return fmt.Errorf("update workout: %w", err)
		}
		return nil
	})
}

func (s *Service) UpdateWorkoutByFilter(ctx context.Context, f dto.WorkoutsFilter, params entities.WorkoutUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		workout, err := s.workoutsRepository.Get(ctx, f, true) // withBlock для блокировки записи
		if err != nil {
			return fmt.Errorf("update workout: get workout: %w", err)
		}

		workout.Update(params)

		if err := s.workoutsRepository.Update(ctx, workout); err != nil {
			return fmt.Errorf("update workout: save: %w", err)
		}

		return nil
	})
}

func (s *Service) DeleteWorkout(ctx context.Context, f dto.WorkoutsFilter) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		exercisesFilter := dto.WorkoutsExerciseFilter{
			WorkoutID: f.ID,
		}

		exercises, err := s.workoutsExerciseRepository.List(ctx, exercisesFilter, true)
		if err != nil {
			return fmt.Errorf("delete workout: list exercises: %w", err)
		}

		for _, exercise := range exercises {
			exID := exercise.ExerciseID()
			exerciseFilter := dto.WorkoutsExerciseFilter{
				WorkoutID:  f.ID,
				ExerciseID: &exID,
			}
			if err := s.workoutsExerciseRepository.Delete(ctx, exerciseFilter); err != nil {
				return fmt.Errorf("delete workout: delete exercise: %w", err)
			}
		}

		if err := s.workoutsRepository.Delete(ctx, f); err != nil {
			return fmt.Errorf("delete workout: delete: %w", err)
		}

		return nil
	})
}

func (s *Service) GetWorkoutExercise(ctx context.Context, f dto.WorkoutsExerciseFilter, withBlock bool) (*entities.WorkoutsExercise, error) {
	workoutExercise, err := s.workoutsExerciseRepository.Get(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("get workout exercise: %w", err)
	}
	return workoutExercise, nil
}

func (s *Service) ListWorkoutsExercise(ctx context.Context, f dto.WorkoutsExerciseFilter) ([]*entities.WorkoutsExercise, error) {
	exercises, err := s.workoutsExerciseRepository.List(ctx, f, false)
	if err != nil {
		return nil, fmt.Errorf("list workout exercises: %w", err)
	}
	return exercises, nil
}

func (s *Service) CreateWorkoutExercise(ctx context.Context, workoutExercise *entities.WorkoutsExercise) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.workoutsExerciseRepository.Create(ctx, workoutExercise); err != nil {
			return fmt.Errorf("create workout exercise: %w", err)
		}
		return nil
	})
}

func (s *Service) CreateWorkoutExercises(ctx context.Context, exercises []*entities.WorkoutsExercise) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		for _, exercise := range exercises {
			if err := s.workoutsExerciseRepository.Create(ctx, exercise); err != nil {
				return fmt.Errorf("create workout exercise: %w", err)
			}
		}
		return nil
	})
}

func (s *Service) UpdateWorkoutExercise(ctx context.Context, workoutExercise *entities.WorkoutsExercise) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.workoutsExerciseRepository.Update(ctx, workoutExercise); err != nil {
			return fmt.Errorf("update workout exercise: %w", err)
		}
		return nil
	})
}

func (s *Service) UpdateWorkoutExerciseByFilter(ctx context.Context, f dto.WorkoutsExerciseFilter, params entities.WorkoutsExerciseUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		// Получаем упражнение по фильтру
		workoutExercise, err := s.workoutsExerciseRepository.Get(ctx, f, true)
		if err != nil {
			return fmt.Errorf("update workout exercise: get: %w", err)
		}

		// Обновляем поля
		workoutExercise.Update(params)

		// Сохраняем изменения
		if err := s.workoutsExerciseRepository.Update(ctx, workoutExercise); err != nil {
			return fmt.Errorf("update workout exercise: save: %w", err)
		}

		return nil
	})
}

func (s *Service) DeleteWorkoutExercise(ctx context.Context, f dto.WorkoutsExerciseFilter) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.workoutsExerciseRepository.Delete(ctx, f); err != nil {
			return fmt.Errorf("delete workout exercise: %w", err)
		}
		return nil
	})
}

func (s *Service) ListExercises(ctx context.Context, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error) {
	exercises, err := s.exercisesRepository.List(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	return exercises, nil
}

func (s *Service) ListExercisesByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Exercise, error) {
	if len(ids) == 0 {
		return []*entities.Exercise{}, nil
	}

	// Создаем фильтр для поиска по нескольким ID
	// Здесь нужно реализовать логику поиска по списку ID
	// Можно либо вызывать Get для каждого ID, либо добавить метод в репозиторий

	exercises := make([]*entities.Exercise, 0, len(ids))
	for _, id := range ids {
		filter := dto.ExerciseFilter{
			ID: &id,
		}
		exercise, err := s.exercisesRepository.Get(ctx, filter, false)
		if err != nil {
			// Пропускаем не найденные упражнения
			continue
		}
		exercises = append(exercises, exercise)
	}

	return exercises, nil
}

func (s *Service) UpdateExerciseByFilter(ctx context.Context, f dto.ExerciseFilter, params entities.ExerciseUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		exercise, err := s.exercisesRepository.Get(ctx, f, true)
		if err != nil {
			return fmt.Errorf("update exercise: get: %w", err)
		}

		exercise.Update(params)

		if err := s.exercisesRepository.Update(ctx, exercise); err != nil {
			return fmt.Errorf("update exercise: save: %w", err)
		}

		return nil
	})
}
