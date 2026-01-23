package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
)

func (a *API) registerCRUDHandlers(router *gin.RouterGroup) {
	group := router.Group("/crud")

	user := group.Group("/user")
	user.DELETE("/info", a.deleteUserInfo)
	user.PATCH("/info", a.updateUserInfo)
	user.GET("/info", a.getUserInfo)

	user.GET("/params", a.getUserParams)
	user.PATCH("/params", a.updateUserParams)
	user.DELETE("/params", a.deleteUserParams)
	user.POST("/params", a.createUserParams)

	user.GET("/weight", a.getUserWeight)
	user.GET("/weight/history", a.getUserWeightHistory)
	user.PATCH("/weight", a.updateUserWeight)
	user.DELETE("/weight", a.deleteUserWeight)
	user.POST("/weight", a.createUserWeight)
}

// deleteUserInfo удаляет информацию о пользователе
// @Summary Удаление информации пользователя
// @Description Удаляет основную информацию о пользователе
// @Tags User Info
// @Security BearerAuth
// @Produce json
// @Success 200 {string} string "Успешное удаление"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/info [delete]
func (a *API) deleteUserInfo(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: delete user info: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: delete user info: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: delete user info: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	err = a.CRUDService.DeleteInfoUser(ctx, dto.UserInfoFilter{ID: &userID})
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: delete user info: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}

// updateUserInfo обновляет информацию о пользователе
// @Summary Обновление информации пользователя
// @Description Обновляет основную информацию о пользователе
// @Tags User Info
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UserInfoUpdateRequestModel true "Данные для обновления"
// @Success 200 {string} map[string]string "Успешное обновление"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/info [patch]
func (a *API) updateUserInfo(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: update user info: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: update user info: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: update user info: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserInfoUpdateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "update user info")
		return
	}

	err = a.CRUDService.UpdateInfoUser(ctx, dto.UserInfoFilter{ID: &userID}, m.ToParam())
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: update user info: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully updated"})
}

// getUserInfo получает информацию о пользователе
// @Summary Получение информации пользователя
// @Description Получает основную информацию о пользователе
// @Tags User Info
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserInfoResponseModel "Информация о пользователе"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/info [get]
func (a *API) getUserInfo(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: get user info: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: get user info: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: get user info: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	ui, err := a.CRUDService.GetInfoUser(ctx, dto.UserInfoFilter{ID: &userID}, false)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: get user info: success")
	ctx.JSON(http.StatusOK, models.NewUserInfoResponse(ui))
}

// getUserParams получает параметры пользователя
// @Summary Получение параметров пользователя
// @Description Получает дополнительные параметры пользователя (рост, возраст и т.д.)
// @Tags User Params
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserParamsResponseModel "Параметры пользователя"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/params [get]
func (a *API) getUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: get user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: get user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: get user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	up, err := a.CRUDService.GetParamsUser(ctx, dto.UserParamsFilter{UserID: &userID}, false)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: get user params: success")
	ctx.JSON(http.StatusOK, models.NewUserParamsResponse(up))
}

// updateUserParams обновляет параметры пользователя
// @Summary Обновление параметров пользователя
// @Description Обновляет дополнительные параметры пользователя (рост, возраст и т.д.)
// @Tags User Params
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UserParamsUpdateRequestModel true "Данные параметров для обновления"
// @Success 200 {object} map[string]string "Успешное обновление"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/params [patch]
func (a *API) updateUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: update user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: update user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: update user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserParamsUpdateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "update user params")
		return
	}

	spec, err := m.ToParam()
	if err != nil {
		a.log.Errorf("crud error: invalid data: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: invalid data": err.Error()})
		return
	}

	err = a.CRUDService.UpdateParamsUser(ctx, dto.UserParamsFilter{UserID: &userID}, spec)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: update user params: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully updated"})
}

func (a *API) handleValidationErrors(c *gin.Context, err error, contextKey string) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		a.log.Errorf("crud error: %s: %s", contextKey, err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error": "Internal validation error"})
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
			out[field] = field + " is too small"
		case "max":
			out[field] = field + " is too big"
		case "oneof":
			out[field] = field + " is invalid"
		default:
			out[field] = field + " is invalid"
		}
	}

	response := gin.H{
		"crud error": gin.H{
			contextKey: out,
		},
	}
	a.log.Errorf("crud error: validation error: %s", response)
	c.AbortWithStatusJSON(http.StatusBadRequest, response)
}

// deleteUserParams удаляет параметры пользователя
// @Summary Удаление параметров пользователя
// @Description Удаляет дополнительные параметры пользователя
// @Tags User Params
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string "Успешное удаление"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/params [delete]
func (a *API) deleteUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: delete user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: delete user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: delete user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	err = a.CRUDService.DeleteParamsUser(ctx, dto.UserParamsFilter{UserID: &userID})
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: delete user params: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}

// createUserParams создает параметры пользователя
// @Summary Создание параметров пользователя
// @Description Создает дополнительные параметры пользователя (рост, возраст и т.д.)
// @Tags User Params
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UserParamsCreateRequestModel true "Данные параметров для создания"
// @Success 200 {object} map[string]string "Успешное создание"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/params [post]
func (a *API) createUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: create user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: create user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: create user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserParamsCreateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "create user params")
		return
	}
	up, err := m.ToSpec()
	if err != nil {
		a.log.Errorf("crud error: create user params: invalid data: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: create user params: invalid data": err.Error()})
		return
	}
	up.UserID = userID

	if err := a.CRUDService.CreateParamsUser(ctx, up); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: create user params: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully created"})
}

// getUserWeight получает текущий вес пользователя
// @Summary Получение текущего веса пользователя
// @Description Получает текущий вес пользователя
// @Tags User Weight
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserWeightResponseModel "Актуальный (последний) вес пользователя"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/weight [get]
func (a *API) getUserWeight(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: get user weight: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: get user weight: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: get user weight: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	uw, err := a.CRUDService.GetWeightUser(ctx, dto.UserWeightFilter{UserID: &userID}, false)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: get user weight: success")
	ctx.JSON(http.StatusOK, models.NewUserWeightResponse(uw))
}

// getUserWeightHistory получает историю веса пользователя
// @Summary Получение истории веса пользователя
// @Description Получает всю историю изменений веса пользователя
// @Tags User Weight
// @Security BearerAuth
// @Produce json
// @Success 200 {object} []models.UserWeightResponseModel "История веса пользователя"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/weight/history [get]
func (a *API) getUserWeightHistory(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: get user weight: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: get user weight history: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: get user weight history: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	uwl, err := a.CRUDService.ListWeightsUser(ctx, dto.UserWeightFilter{UserID: &userID}, false)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: get user weight history: success")
	ctx.JSON(http.StatusOK, models.NewUserWeightResponseList(uwl))
}

// updateUserWeight обновляет вес пользователя
// @Summary Обновление веса пользователя (пока нет)
// @Description Обновляет текущий вес пользователя
// @Tags User Weight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {string} map[string]string "Успешное обновление"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/weight [patch]
func (a *API) updateUserWeight(ctx *gin.Context) {}

// deleteUserWeight удаляет запись о весе пользователя по айди записи (пока нет)
// @Summary Удаление записи о весе пользователя
// @Description Удаляет запись о весе пользователя по ID
// @Tags User Weight
// @Security BearerAuth
// @Produce json
// @Param id query string true "ID записи о весе"
// @Success 200 {string} map[string]string "Успешное удаление"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/weight [delete]
func (a *API) deleteUserWeight(ctx *gin.Context) {}

// createUserWeight создает запись о весе пользователя
// @Summary Создание записи о весе пользователя
// @Description Создает новую запись о весе пользователя
// @Tags User Weight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UserWeightCreateRequestModel true "Данные веса для создания"
// @Success 200 {string} map[string]string "Успешное создание"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} map[string]interface{} "Отсутствует авторизация"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /crud/user/weight [post]
func (a *API) createUserWeight(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: create user weight: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: create user weight: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: create user weight: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserWeightCreateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "create user weight")
		return
	}
	uw := m.ToSpec()
	if err != nil {
		a.log.Errorf("crud error: create user weight: invalid data: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: create user weight: invalid data": err.Error()})
		return
	}
	uw.UserID = userID

	if err := a.CRUDService.CreateWeightUser(ctx, uw); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: create user weight: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully created"})
}
