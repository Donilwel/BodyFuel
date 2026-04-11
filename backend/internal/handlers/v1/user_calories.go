package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *API) registerUserCaloriesHandlers(router *gin.RouterGroup) {
	user := router.Group("/user/calories")
	user.GET("/history", a.getUserCaloriesHistory)
	user.POST("", a.createUserCalories)
	user.PATCH("/:uuid", a.updateUserCalories)
	user.DELETE("/:uuid", a.deleteUserCalories)
}

// getUserCaloriesHistory возвращает историю калорий пользователя
// @Summary История записей о калориях
// @Description Возвращает список записей о потреблённых/затраченных калориях за период
// @Tags User Calories
// @Security BearerAuth
// @Produce json
// @Param start_date query string false "Начало периода (ISO 8601)"
// @Param end_date query string false "Конец периода (ISO 8601)"
// @Success 200 {array} models.UserCaloriesResponse "История калорий"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/history [get]
func (a *API) getUserCaloriesHistory(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	f := dto.UserCaloriesFilter{UserID: &userID}

	if s := ctx.Query("start_date"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use RFC3339"})
			return
		}
		f.StartDate = &t
	}
	if e := ctx.Query("end_date"); e != "" {
		t, err := time.Parse(time.RFC3339, e)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use RFC3339"})
			return
		}
		f.EndDate = &t
	}

	list, err := a.CRUDService.ListUserCalories(ctx, f)
	if err != nil {
		a.log.Errorf("user calories: list error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list calories"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewUserCaloriesResponseList(list))
}

// createUserCalories создаёт запись о калориях
// @Summary Создание записи о калориях
// @Description Создаёт новую запись о потреблённых или затраченных калориях
// @Tags User Calories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateUserCaloriesRequest true "Данные о калориях"
// @Success 201 {object} models.SuccessResponse "Запись создана"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories [post]
func (a *API) createUserCalories(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	var m models.CreateUserCaloriesRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "create user calories")
		return
	}

	spec := entities.UserCaloriesInitSpec{
		ID:          uuid.New(),
		UserID:      userID,
		Calories:    m.Calories,
		Description: m.Description,
		Date:        m.Date,
	}

	if err := a.CRUDService.CreateUserCalories(ctx, spec); err != nil {
		a.log.Errorf("user calories: create error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create calories entry"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Successfully created"})
}

// updateUserCalories обновляет запись о калориях
// @Summary Обновление записи о калориях
// @Description Обновляет существующую запись о калориях по ID
// @Tags User Calories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uuid path string true "ID записи"
// @Param request body models.UpdateUserCaloriesRequest true "Поля для обновления"
// @Success 204 {object} models.SuccessResponse "Запись обновлена"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/{uuid} [patch]
func (a *API) updateUserCalories(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	entryID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var m models.UpdateUserCaloriesRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "update user calories")
		return
	}

	params := entities.UserCaloriesUpdateParams{
		Calories:    m.Calories,
		Description: m.Description,
		Date:        m.Date,
	}

	f := dto.UserCaloriesFilter{ID: &entryID, UserID: &userID}

	if err := a.CRUDService.UpdateUserCalories(ctx, f, params); err != nil {
		a.log.Errorf("user calories: update error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update calories entry"})
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{"message": "Successfully updated"})
}

// deleteUserCalories удаляет запись о калориях
// @Summary Удаление записи о калориях
// @Description Удаляет запись о калориях по ID
// @Tags User Calories
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID записи"
// @Success 204 {object} models.SuccessResponse "Запись удалена"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/calories/{uuid} [delete]
func (a *API) deleteUserCalories(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	entryID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	if err := a.CRUDService.DeleteUserCalories(ctx, entryID, userID); err != nil {
		a.log.Errorf("user calories: delete error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete calories entry"})
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{"message": "Successfully deleted"})
}
