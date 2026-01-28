package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// getUserParams получает параметры пользователя
// @Summary Получение параметров пользователя
// @Description Получает дополнительные параметры пользователя (рост, возраст и т.д.)
// @Tags User Params
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserParamsResponseModel "Параметры пользователя"
// @Failure 400 {object} string "Неверный формат ID"
// @Failure 401 {object} string "Отсутствует авторизация"
// @Failure 500 {object} string "Внутренняя ошибка сервера"
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
// @Success 200 {object} string "Успешное обновление"
// @Failure 400 {object} string "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} string "Отсутствует авторизация"
// @Failure 500 {object} string "Внутренняя ошибка сервера"
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

// deleteUserParams удаляет параметры пользователя
// @Summary Удаление параметров пользователя
// @Description Удаляет дополнительные параметры пользователя
// @Tags User Params
// @Security BearerAuth
// @Produce json
// @Success 200 {object} string "Успешное удаление"
// @Failure 400 {object} string "Неверный формат ID"
// @Failure 401 {object} string "Отсутствует авторизация"
// @Failure 500 {object} string "Внутренняя ошибка сервера"
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
// @Success 200 {object} string "Успешное создание"
// @Failure 400 {object} string "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} string "Отсутствует авторизация"
// @Failure 500 {object} string "Внутренняя ошибка сервера"
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
