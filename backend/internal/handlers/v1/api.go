package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/JWT"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type (
	AuthService interface {
		Register(ctx context.Context, ua entities.UserInfoInitSpec) error
		Login(ctx context.Context, ui entities.UserAuthInitSpec) (string, error)
	}

	UserStatisticsService interface {
		CreateUserParams(ctx context.Context, up entities.UserParamsInitSpec) error
	}

	CRUDService interface {
		GetInfoUser(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
		CreateInfoUser(ctx context.Context, info entities.UserInfoInitSpec) error
		UpdateInfoUser(ctx context.Context, f dto.UserInfoFilter, info entities.UserInfoUpdateParams) error
		DeleteInfoUser(ctx context.Context, f dto.UserInfoFilter) error

		GetParamsUser(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error)
		CreateParamsUser(ctx context.Context, params entities.UserParamsInitSpec) error
		UpdateParamsUser(ctx context.Context, f dto.UserParamsFilter, userParams entities.UserParamsUpdateParams) error
		DeleteParamsUser(ctx context.Context, f dto.UserParamsFilter) error
	}
)

type Config struct {
	AuthService           AuthService
	UserStatisticsService UserStatisticsService
	CRUDService           CRUDService
	Validator             validator.Validate
}

type API struct {
	authService           AuthService
	userStatisticsService UserStatisticsService
	CRUDService           CRUDService
	validator             validator.Validate
}

func NewHandlers(c Config) *API {
	return &API{
		authService:           c.AuthService,
		userStatisticsService: c.UserStatisticsService,
		CRUDService:           c.CRUDService,
		validator:             c.Validator,
	}
}

func (a *API) RegisterHandlers(r *gin.RouterGroup) {
	a.registerAuthHandlers(r)

	protected := r.Group("", JWT.JWTAuthMiddleware())
	a.registerParamsHandlers(r)
	a.registerCRUDHandlers(protected)

}
