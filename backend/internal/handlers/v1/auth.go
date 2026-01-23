package v1

import (
	"backend/internal/handlers/v1/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"net/http"
	"regexp"
)

func (a *API) registerAuthHandlers(router *gin.RouterGroup) {
	group := router.Group("/auth")
	group.POST("/register", a.register)
	group.POST("/login", a.login)
}

// register обрабатывает регистрацию пользователя
// @Summary Регистрация нового пользователя
// @Description Создание нового аккаунта пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequestModel true "Данные для регистрации"
// @Success 201 {string}	string	"Пользователь успешно зарегистрирован"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации"
// @Failure 409 {object} map[string]interface{} "Пользователь уже существует"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /auth/register [post]
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

	err := a.checkPhone(ctx, r.Phone)
	if err != nil {
		a.log.Errorf("%s: %v", "auth error", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
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

func (a *API) checkPhone(ctx *gin.Context, phone string) error {
	phoneRegex := `^\+?[0-9]{10,15}$`
	re := regexp.MustCompile(phoneRegex)

	if !re.MatchString(phone) {
		return fmt.Errorf("uncorrect inpit phone number")
	}

	return nil
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

// login обрабатывает вход пользователя
// @Summary Аутентификация пользователя
// @Description Вход пользователя в систему и получение JWT токена
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequestModel true "Данные для входа"
// @Success 200 {object} models.JWTModel "Успешная аутентификация"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации"
// @Failure 401 {object} map[string]interface{} "Неверные учетные данные"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /auth/login [post]
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
