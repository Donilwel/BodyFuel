package v1

import (
	"backend/internal/handlers/v1/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *API) registerAuthHandlers(router *gin.RouterGroup) {
	group := router.Group("/auth")
	group.POST("/register", a.register)
	group.GET("/login", a.login)
}

func (a *API) register(ctx *gin.Context) {
	var r models.RegisterRequestModel
	if err := ctx.ShouldBindJSON(&r); err != nil {
		ctx.Error(err)
		return
	}
	if err := a.authService.Register(ctx, r.ToSpec()); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Successfully registers"})
}

func (a *API) login(ctx *gin.Context) {
	var m models.LoginRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.Error(err)
		return
	}

	jwt, err := a.authService.Login(ctx, m.ToSpec())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, models.NewJWTCodeResponse(jwt))
}
