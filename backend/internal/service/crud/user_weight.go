package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"fmt"
)

func (s *Service) GetWeightUser(ctx context.Context, f dto.UserWeightFilter, withBlock bool) (*entities.UserWeight, error) {
	uw, err := s.userWeightRepository.Get(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("get user weight: %w", err)
	}
	return uw, nil
}

func (s *Service) ListWeightsUser(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error) {
	uw, err := s.userWeightRepository.List(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("list user weight: %w", err)
	}

	return uw, nil
}

func (s *Service) CreateWeightUser(ctx context.Context, weight entities.UserWeightInitSpec) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.userWeightRepository.Create(ctx, entities.NewUserWeight(entities.WithUserWeightInitSpec(weight))); err != nil {
			return fmt.Errorf("create user weight: %w", err)
		}
		return nil
	})
}

func (s *Service) UpdateWeightUser(ctx context.Context, f dto.UserWeightFilter, weight entities.UserWeightUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		uw, err := s.userWeightRepository.Get(ctx, f, false)
		if err != nil {
			return fmt.Errorf("update user weight: get user weight: %w", err)
		}
		uw.Update(weight)

		if err := s.userWeightRepository.Update(ctx, uw); err != nil {
			return fmt.Errorf("update user weight: update: %w", err)
		}
		return nil
	})
}

func (s *Service) DeleteWeightUser(ctx context.Context, f dto.UserWeightFilter) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.userWeightRepository.Delete(ctx, f); err != nil {
			return fmt.Errorf("delete user weight: delete:%w", err)
		}
		return nil
	})
}
