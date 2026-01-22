package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"context"
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

	UserWeightRepository interface {
		Get(ctx context.Context, f dto.UserWeightFilter, withBlock bool) (*entities.UserWeight, error)
		Create(ctx context.Context, userWeight *entities.UserWeight) error
		Update(ctx context.Context, userWeight *entities.UserWeight) error
		Delete(ctx context.Context, f dto.UserWeightFilter) error
		List(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error)
	}

	TransactionManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	}
)

type Config struct {
	TransactionManager   TransactionManager
	UserInfoRepository   UserInfoRepository
	UserParamsRepository UserParamsRepository
	UserWeightRepository UserWeightRepository
	Log                  logging.Entry
}

type Service struct {
	transactionManager   TransactionManager
	userInfoRepository   UserInfoRepository
	userParamsRepository UserParamsRepository
	userWeightRepository UserWeightRepository
	log                  logging.Entry
}

func NewService(c *Config) *Service {
	return &Service{
		transactionManager:   c.TransactionManager,
		userInfoRepository:   c.UserInfoRepository,
		userParamsRepository: c.UserParamsRepository,
		userWeightRepository: c.UserWeightRepository,
		log:                  c.Log,
	}
}
