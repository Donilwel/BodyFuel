package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/handlers/v1/models"
	"backend/pkg/JWT"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"net/http"
)

func (a *API) registerAuthHandlers(router *gin.RouterGroup) {
	group := router.Group("/auth")
	group.POST("/register", a.register)
	group.POST("/login", a.login)
	group.POST("/refresh", a.refresh)
	group.POST("/recover", a.sendRecoveryCode)
	group.POST("/reset-password", a.resetPassword)

	protected := group.Group("", JWT.JWTAuthMiddleware())
	protected.POST("/verify-email", a.verifyEmail)
	protected.POST("/verify-phone", a.verifyPhone)
	protected.POST("/send-verification", a.sendVerificationCode)
}

// register обрабатывает регистрацию пользователя
// @Summary Регистрация нового пользователя
// @Description Создание нового аккаунта пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequestModel true "Данные для регистрации"
// @Success 201 {object} models.SuccessResponse	"Пользователь успешно зарегистрирован"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 409 {object} models.ErrorResponse "Пользователь уже существует"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
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

// login обрабатывает вход пользователя
// @Summary Аутентификация пользователя
// @Description Вход пользователя в систему и получение пары токенов
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequestModel true "Данные для входа"
// @Success 200 {object} models.TokenPairModel "Успешная аутентификация"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
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

	pair, err := a.authService.Login(ctx, m.ToSpec())
	if err != nil {
		a.log.Errorf("%s: %v", "auth error", err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"auth error": err.Error()})
		return
	}
	a.log.Info("auth: login: success")
	ctx.JSON(http.StatusOK, models.NewTokenPairModel(pair))
}

// refresh обновляет пару токенов по refresh token
// @Summary Обновление токенов
// @Description Получение новой пары access+refresh токенов
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.TokenPairModel "Новая пара токенов"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Недействительный токен"
// @Router /auth/refresh [post]
func (a *API) refresh(ctx *gin.Context) {
	var m models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "refresh")
		return
	}

	pair, err := a.authService.Refresh(ctx, m.RefreshToken)
	if err != nil {
		a.log.Errorf("auth: refresh: %v", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"auth error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, models.NewTokenPairModel(pair))
}

// sendVerificationCode отправляет код подтверждения
// @Summary Отправка кода подтверждения
// @Description Отправка 6-значного кода на email или phone
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SendVerificationCodeRequest true "Тип кода"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/send-verification [post]
func (a *API) sendVerificationCode(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("exercise error: get exercise: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("exercise error: get user exercise: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("exercise error: get user exercise: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.SendVerificationCodeRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "send-verification")
		return
	}

	codeType := entities.VerificationCodeType(m.CodeType)
	if err := a.authService.SendVerificationCode(ctx, userID, codeType); err != nil {
		a.log.Errorf("auth: send verification code: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"auth error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
}

// verifyEmail подтверждает email пользователя
// @Summary Подтверждение email
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.VerifyCodeRequest true "Код подтверждения"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/verify-email [post]
func (a *API) verifyEmail(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("exercise error: get exercise: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("exercise error: get user exercise: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("exercise error: get user exercise: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}
	
	var m models.VerifyCodeRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "verify-email")
		return
	}

	if err := a.authService.VerifyCode(ctx, userID, m.Code, entities.VerificationCodeEmail); err != nil {
		a.log.Errorf("auth: verify email: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Email verified"})
}

// verifyPhone подтверждает телефон пользователя
// @Summary Подтверждение телефона
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.VerifyCodeRequest true "Код подтверждения"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/verify-phone [post]
func (a *API) verifyPhone(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	var m models.VerifyCodeRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "verify-phone")
		return
	}

	if err := a.authService.VerifyCode(ctx, userID, m.Code, entities.VerificationCodePhone); err != nil {
		a.log.Errorf("auth: verify phone: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Phone verified"})
}

// sendRecoveryCode отправляет код для восстановления пароля
// @Summary Отправка кода восстановления пароля
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RecoverPasswordRequest true "Email"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/recover [post]
func (a *API) sendRecoveryCode(ctx *gin.Context) {
	var m models.RecoverPasswordRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "recover")
		return
	}

	// Always respond 200 to avoid user enumeration
	_ = a.authService.SendRecoveryCode(ctx, m.Email)
	ctx.JSON(http.StatusOK, gin.H{"message": "If the email exists, a recovery code has been sent"})
}

// resetPassword сбрасывает пароль по коду
// @Summary Сброс пароля
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Данные для сброса"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/reset-password [post]
func (a *API) resetPassword(ctx *gin.Context) {
	var m models.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "reset-password")
		return
	}

	if err := a.authService.ResetPassword(ctx, m.Email, m.Code, m.NewPassword); err != nil {
		a.log.Errorf("auth: reset password: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"auth error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
