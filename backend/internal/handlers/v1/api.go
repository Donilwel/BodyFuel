package v1

import (
	"backend/internal/domain/entities"
	"context"
	"github.com/gin-gonic/gin"
)

type (
	CRUDService interface {
	}
	AuthService interface {
		Register(ctx context.Context, ua entities.UserInfoInitSpec) error
		Login(ctx context.Context, ui entities.UserAuthInitSpec) (string, error)
	}
)

type Config struct {
	AuthService AuthService
}

type API struct {
	authService AuthService
}

func NewHandlers(c Config) *API {
	return &API{
		authService: c.AuthService,
	}
}

func (a *API) RegisterHandlers(r *gin.RouterGroup) {
	a.registerAuthHandlers(r)
}
