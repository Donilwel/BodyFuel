package params

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"context"
)

type (
	UserParamsRepository interface {
		Get(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error)
		Create(ctx context.Context, userParams *entities.UserParams) error
		Update(ctx context.Context, userParams *entities.UserParams) error
		Delete(ctx context.Context, f dto.UserParamsFilter) error
	}

	TransactionManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	}

	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
	}
)

type Service struct {
	txm            TransactionManager
	userParamsRepo UserParamsRepository
	userInfoRepo   UserInfoRepository
	log            logging.Entry
}

type Config struct {
	TransactionManager   TransactionManager
	UserParamsRepository UserParamsRepository
	UserInfoRepository   UserInfoRepository
	Log                  logging.Entry
}

func NewService(c *Config) *Service {
	return &Service{
		txm:            c.TransactionManager,
		userParamsRepo: c.UserParamsRepository,
		userInfoRepo:   c.UserInfoRepository,
		log:            c.Log,
	}
}

func (u *Service) Get(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error) {
	return u.userParamsRepo.Get(ctx, f, withBlock)
}
