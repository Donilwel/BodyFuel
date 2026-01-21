package v1

import (
	"backend/internal/domain/entities"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type (
	AuthService interface {
		Register(ctx context.Context, ua entities.UserInfoInitSpec) error
		Login(ctx context.Context, ui entities.UserAuthInitSpec) (string, error)
	}
)

type Config struct {
	AuthService AuthService
	Validator   validator.Validate
}

type API struct {
	authService AuthService
	validator   validator.Validate
}

func NewHandlers(c Config) *API {
	return &API{
		authService: c.AuthService,
		validator:   c.Validator,
	}
}

func (a *API) RegisterHandlers(r *gin.RouterGroup) {
	a.registerAuthHandlers(r)
}
