package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/errors"
	"context"
	"fmt"
)

func (s *Service) GetParamsUser(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error) {
	user, err := s.userParamsRepository.Get(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("get user params: %w", err)
	}
	return user, nil
}

func (s *Service) CreateParamsUser(ctx context.Context, params entities.UserParamsInitSpec) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if _, err := s.userParamsRepository.Get(ctx, dto.UserParamsFilter{UserID: &params.UserID}, false); err == nil {
			return fmt.Errorf("create user params: %w", errors.ErrUserParamsAlreadyExists)
		}

		if err := s.userParamsRepository.Create(ctx, entities.NewUserParams(entities.WithUserParamsInitSpec(params))); err != nil {
			return fmt.Errorf("create user params: %w", err)
		}
		return nil
	})
}

func (s *Service) UpdateParamsUser(ctx context.Context, f dto.UserParamsFilter, userParams entities.UserParamsUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		up, err := s.userParamsRepository.Get(ctx, f, false)
		if err != nil {
			return fmt.Errorf("update user params: get user params: %w", err)
		}
		up.Update(userParams)

		if err := s.userParamsRepository.Update(ctx, up); err != nil {
			return fmt.Errorf("update user params: update: %w", err)
		}
		return nil
	})
}

func (s *Service) DeleteParamsUser(ctx context.Context, f dto.UserParamsFilter) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.userParamsRepository.Delete(ctx, f); err != nil {
			return fmt.Errorf("delete user params: delete:%w", err)
		}
		return nil
	})
}
