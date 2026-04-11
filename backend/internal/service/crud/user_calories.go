package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateUserCalories(ctx context.Context, spec entities.UserCaloriesInitSpec) error {
	uc := entities.NewUserCalories(entities.WithUserCaloriesInitSpec(spec))
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.userCaloriesRepository.Create(ctx, uc); err != nil {
			return fmt.Errorf("create user calories: %w", err)
		}
		return nil
	})
}

func (s *Service) GetUserCalories(ctx context.Context, f dto.UserCaloriesFilter) (*entities.UserCalories, error) {
	uc, err := s.userCaloriesRepository.Get(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("get user calories: %w", err)
	}
	return uc, nil
}

func (s *Service) ListUserCalories(ctx context.Context, f dto.UserCaloriesFilter) ([]*entities.UserCalories, error) {
	list, err := s.userCaloriesRepository.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list user calories: %w", err)
	}
	return list, nil
}

func (s *Service) UpdateUserCalories(ctx context.Context, f dto.UserCaloriesFilter, params entities.UserCaloriesUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		uc, err := s.userCaloriesRepository.Get(ctx, f)
		if err != nil {
			return fmt.Errorf("update user calories: get: %w", err)
		}

		uc.Update(params)

		if err := s.userCaloriesRepository.Update(ctx, uc); err != nil {
			return fmt.Errorf("update user calories: save: %w", err)
		}
		return nil
	})
}

func (s *Service) DeleteUserCalories(ctx context.Context, id, userID uuid.UUID) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.userCaloriesRepository.Delete(ctx, dto.UserCaloriesFilter{
			ID:     &id,
			UserID: &userID,
		}); err != nil {
			return fmt.Errorf("delete user calories: %w", err)
		}
		return nil
	})
}

// ListUserCaloriesForPeriod — удобная обёртка с датами.
func (s *Service) ListUserCaloriesForPeriod(ctx context.Context, userID uuid.UUID, start, end *time.Time) ([]*entities.UserCalories, error) {
	return s.ListUserCalories(ctx, dto.UserCaloriesFilter{
		UserID:    &userID,
		StartDate: start,
		EndDate:   end,
	})
}
