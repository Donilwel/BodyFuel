package v1

import (
	"backend/internal/handlers/v1/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

func (a *API) registerAuthHandlers(router *gin.RouterGroup) {
	group := router.Group("/auth")
	group.POST("/register", a.register)
	group.POST("/login", a.login)
}

func (a *API) register(ctx *gin.Context) {
	var r models.RegisterRequestModel
	if err := ctx.ShouldBindJSON(&r); err != nil {
		a.log.Errorf("%s: %v", "auth error", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(r); err != nil {
		a.handleValidationAuthFields(ctx, err, "register")
		return
	}

	if err := a.authService.Register(ctx, r.ToSpec()); err != nil {
		a.log.Errorf("%s: %v", "auth error", err.Error())
		ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"auth error": err.Error()})
		return
	}
	a.log.Info("auth: register: success")
	ctx.JSON(http.StatusCreated, gin.H{"message": "Successfully registers"})
}

func (a *API) handleValidationAuthFields(c *gin.Context, err error, typeMethod string) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		a.log.Errorf("%s: %v", "auth error", "Internal error")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"auth error": "Internal error"})
		return
	}

	out := make(map[string]string)
	for _, fe := range ve {
		field := fe.Field()
		tag := fe.Tag()

		switch tag {
		case "required":
			out[field] = field + " is required"
		case "min":
			out[field] = field + " is too short"
		case "regex":
			out[field] = "phone number is invalid"
		case "email":
			out[field] = "email is invalid"
		default:
			out[field] = field + " is invalid"
		}
	}

	response := gin.H{
		"auth error": gin.H{
			typeMethod: out,
		},
	}
	a.log.Errorf("%s: %s: %v", "auth error", typeMethod, out)
	c.AbortWithStatusJSON(http.StatusBadRequest, response)
}

func (a *API) login(ctx *gin.Context) {
	var m models.LoginRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("%s: %v", "auth error", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "login")
		return
	}

	jwt, err := a.authService.Login(ctx, m.ToSpec())
	if err != nil {
		a.log.Errorf("%s: %v", "auth error", err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"auth error": err.Error()})
		return
	}
	a.log.Info("auth: login: success")
	ctx.JSON(http.StatusOK, models.NewJWTCodeResponse(jwt))
}
