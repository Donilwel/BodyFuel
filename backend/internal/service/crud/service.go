package crud

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"context"
	"github.com/google/uuid"
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

	TasksRepository interface {
		List(ctx context.Context, f dto.TasksFilter, withBlock bool) ([]*entities.Task, error)
		Get(ctx context.Context, f dto.TasksFilter, withBlock bool) (*entities.Task, error)
		Update(ctx context.Context, t *entities.Task) error
		Delete(ctx context.Context, ids []uuid.UUID) error
	}

	ExercisesRepository interface {
		Get(ctx context.Context, f dto.ExerciseFilter, withBlock bool) (*entities.Exercise, error)
		Create(ctx context.Context, exercise *entities.Exercise) error
		Update(ctx context.Context, exercise *entities.Exercise) error
		Delete(ctx context.Context, f dto.ExerciseFilter) error
		List(ctx context.Context, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error)
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
	TasksRepository      TasksRepository
	ExercisesRepository  ExercisesRepository
	Log                  logging.Entry
}

type Service struct {
	transactionManager   TransactionManager
	userInfoRepository   UserInfoRepository
	userParamsRepository UserParamsRepository
	userWeightRepository UserWeightRepository
	tasksRepository      TasksRepository
	exercisesRepository  ExercisesRepository
	log                  logging.Entry
}

func NewService(c *Config) *Service {
	return &Service{
		transactionManager:   c.TransactionManager,
		userInfoRepository:   c.UserInfoRepository,
		userParamsRepository: c.UserParamsRepository,
		userWeightRepository: c.UserWeightRepository,
		tasksRepository:      c.TasksRepository,
		exercisesRepository:  c.ExercisesRepository,
		log:                  c.Log,
	}
}
