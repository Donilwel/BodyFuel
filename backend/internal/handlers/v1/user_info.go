package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// deleteUserInfo удаляет информацию о пользователе
// @Summary Удаление информации пользователя
// @Description Удаляет основную информацию о пользователе
// @Tags User Info
// @Security BearerAuth
// @Produce json
// @Success 200 {string} string "Успешное удаление"
// @Failure 400 {object} string "Неверный формат ID"
// @Failure 401 {object} string "Отсутствует авторизация"
// @Failure 500 {object} string "Внутренняя ошибка сервера"
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
// @Success 200 {string} string "Успешное обновление"
// @Failure 400 {object} string "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} string "Отсутствует авторизация"
// @Failure 500 {object} string "Внутренняя ошибка сервера"
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
// @Failure 400 {object} string "Неверный формат ID"
// @Failure 401 {object} string "Отсутствует авторизация"
// @Failure 500 {object} string "Внутренняя ошибка сервера"
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
