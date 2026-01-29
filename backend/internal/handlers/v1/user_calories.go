package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func (a *API) registerUserCaloriesHandlers(router *gin.RouterGroup) {
	user := router.Group("/user")
	user.DELETE("/calories/:uuid", a.deleteUserCalories)
	user.PATCH("/calories/:uuid", a.updateUserCalories)
	user.GET("/calories/:uuid", a.getUserCalories)
	user.GET("/calories/history", a.getUserCaloriesHistory)
}

// getUserCalories получает текущий вес пользователя
// @Summary Получение текущего веса пользователя
// @Description Получает текущий вес пользователя
// @Tags User Weight
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserWeightResponseModel "Актуальный (последний) вес пользователя"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/{uuid} [get]
func (a *API) getUserCalories(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("user weight error: get user weight: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("user weight error: get user weight: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("user weight error: get user weight: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	uw, err := a.CRUDService.GetWeightUser(ctx, dto.UserWeightFilter{UserID: &userID}, false)
	if err != nil {
		a.log.Errorf("user weight error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("user weight: get user weight: success")
	ctx.JSON(http.StatusOK, models.NewUserWeightResponse(uw))
}

// getUserCaloriesHistory получает историю веса пользователя
// @Summary Получение истории веса пользователя
// @Description Получает всю историю изменений веса пользователя
// @Tags User Weight
// @Security BearerAuth
// @Produce json
// @Success 200 {array} []models.UserWeightResponseModel "История веса пользователя"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/history [get]
func (a *API) getUserCaloriesHistory(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("user weight error: get user weight: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("user weight error: get user weight history: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("user weight error: get user weight history: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	uwl, err := a.CRUDService.ListWeightsUser(ctx, dto.UserWeightFilter{UserID: &userID}, false)
	if err != nil {
		a.log.Errorf("user weight error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"user weight error: internal error": err.Error()})
		return
	}

	a.log.Infof("user weight: get user weight history: success")
	ctx.JSON(http.StatusOK, models.NewUserWeightResponseList(uwl))
}

// updateUserCalories обновляет вес пользователя
// @Summary Обновление веса пользователя (пока нет)
// @Description Обновляет текущий вес пользователя
// @Tags User Weight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {string} models.SuccessResponse "Успешное обновление"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/{uuid} [patch]
func (a *API) updateUserCalories(ctx *gin.Context) {}

// deleteUserCalories удаляет запись о весе пользователя по айди записи
// @Summary Удаление записи о весе пользователя
// @Description Удаляет запись о весе пользователя по ID
// @Tags User Weight
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID записи о весе"
// @Success 200 {string} models.SuccessResponse "Успешное удаление"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/{uuid} [delete]
func (a *API) deleteUserCalories(ctx *gin.Context) {
	id := ctx.Param("uuid")
	if id == "" {
		a.log.Errorf("user weight error: create user weight: missing weight id in header")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing weight id in header",
		})
		return
	}

	weightID, err := uuid.Parse(id)
	if err != nil {
		a.log.Errorf("user weight error: delete user weight: invalid weight id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid id format",
			"details": err.Error(),
		})
		return
	}

	err = a.CRUDService.DeleteWeightUser(ctx, dto.UserWeightFilter{ID: &weightID})
	if err != nil {
		a.log.Errorf("user weight error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"user weight error: internal error": err.Error()})
		return
	}

	a.log.Infof("user weight: delete user weight: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}

// createUserCalories создает запись о весе пользователя
// @Summary Создание записи о весе пользователя
// @Description Создает новую запись о весе пользователя
// @Tags User Weight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UserWeightCreateRequestModel true "Данные веса для создания"
// @Success 200 {string} models.SuccessResponse "Успешное создание"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/{uuid} [post]
func (a *API) createUserCalories(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("user weight error: create user weight: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("user weight error: create user weight: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("user weight error: create user weight: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserWeightCreateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("user weight error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"user weight error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "create user weight")
		return
	}
	uw := m.ToSpec()
	if err != nil {
		a.log.Errorf("user weight error: create user weight: invalid data: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"user weight error: create user weight: invalid data": err.Error()})
		return
	}
	uw.UserID = userID

	if err := a.CRUDService.CreateWeightUser(ctx, uw); err != nil {
		a.log.Errorf("user weight error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"user weight error: internal error": err.Error()})
		return
	}

	a.log.Infof("user weight info: create user weight: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully created"})
}
