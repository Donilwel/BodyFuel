package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/errors"
	"context"
	"fmt"
)

func (s *Service) GetInfoUser(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error) {
	user, err := s.userInfoRepository.Get(ctx, f, withBlock)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	return user, nil
}

// TODO : Бесполезный сервис, может потом куда то можно будет пристроить
func (s *Service) CreateInfoUser(ctx context.Context, info entities.UserInfoInitSpec) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if _, err := s.userInfoRepository.Get(ctx, dto.UserInfoFilter{ID: &info.ID}, false); err == nil {
			return fmt.Errorf("create user info: %w", errors.ErrUserInfoAlreadyExists)
		}
		if err := s.userInfoRepository.Create(ctx, entities.NewUserInfo(entities.WithUserInfoInitSpec(info))); err != nil {
			return fmt.Errorf("create user info: %w", err)
		}
		return nil
	})
}

func (s *Service) UpdateInfoUser(ctx context.Context, f dto.UserInfoFilter, info entities.UserInfoUpdateParams) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		ui, err := s.userInfoRepository.Get(ctx, f, false)
		if err != nil {
			return fmt.Errorf("update user info: get user info: %w", err)
		}

		ui.Update(info)

		if err := s.userInfoRepository.Update(ctx, ui); err != nil {
			return fmt.Errorf("update user info: %w", err)
		}
		return nil
	})
}

func (s *Service) DeleteInfoUser(ctx context.Context, f dto.UserInfoFilter) error {
	return s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.userInfoRepository.Delete(ctx, f); err != nil {
			return fmt.Errorf("delete user info: %w", err)
		}
		return nil
	})
}
