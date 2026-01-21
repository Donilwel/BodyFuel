package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/errors"
	"backend/pkg/logging"
	"context"
	"fmt"
)

type (
	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
		Create(ctx context.Context, userInfo *entities.UserInfo) error
		Update(ctx context.Context, userInfo *entities.UserInfo) error
		Delete(ctx context.Context, f dto.UserInfoFilter) error
	}

	UserParamsRepository interface {
		Get(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error)
		Create(ctx context.Context, userParams *entities.UserParams) error
		Update(ctx context.Context, userParams *entities.UserParams) error
		Delete(ctx context.Context, f dto.UserParamsFilter) error
	}

	TransactionManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	}
)

type Config struct {
	TransactionManager   TransactionManager
	UserInfoRepository   UserInfoRepository
	UserParamsRepository UserParamsRepository
	Log                  logging.Entry
}

type Service struct {
	transactionManager   TransactionManager
	userInfoRepository   UserInfoRepository
	userParamsRepository UserParamsRepository
	log                  logging.Entry
}

func NewService(c *Config) *Service {
	return &Service{
		transactionManager:   c.TransactionManager,
		userInfoRepository:   c.UserInfoRepository,
		userParamsRepository: c.UserParamsRepository,
		log:                  c.Log,
	}
}

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
