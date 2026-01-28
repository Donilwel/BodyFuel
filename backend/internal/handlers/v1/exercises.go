package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func (a *API) registerExerciseHandlers(router *gin.RouterGroup) {
	exercises := router.Group("/exercises")
	exercises.GET("/:uuid", a.getExercise)
	exercises.DELETE("/:uuid", a.deleteExercise)
	exercises.PATCH("/:uuid", a.updateExercise)
	exercises.POST("/", a.createExercise)
	exercises.GET("", a.getExercises)
}

// getExercise получает конкретное упражнение по ID
// @Summary Получение упражнения по ID
// @Description Получает детальную информацию об упражнении по его ID
// @Tags Exercises
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID упражнения"
// @Success 200 {object} models.ExerciseResponseModel "Детальная информация об упражнении"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 404 {object} models.ErrorResponse "Упражнение не найдено"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /exercises/{uuid} [get]
func (a *API) getExercise(ctx *gin.Context) {
	//userIDRaw, ok := ctx.Get("user_id")
	//if !ok {
	//	a.log.Errorf("user weight error: get user weight: missing user_id in context")
	//	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
	//		"error": "missing user_id in context",
	//	})
	//	return
	//}
	//
	//userIDStr, ok := userIDRaw.(string)
	//if !ok {
	//	a.log.Errorf("user weight error: get user weight: invalid user_id type in context")
	//	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
	//		"error": "invalid user_id type in context",
	//	})
	//	return
	//}
	//
	//userID, err := uuid.Parse(userIDStr)
	//if err != nil {
	//	a.log.Errorf("user weight error: get user weight: invalid user_id format: %s", err.Error())
	//	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
	//		"error":   "invalid user_id format",
	//		"details": err.Error(),
	//	})
	//	return
	//}
	//
	//uw, err := a.CRUDService.GetExercise(ctx, dto.ExerciseFilter{ID: id}, false)
	//if err != nil {
	//	a.log.Errorf("user weight error: internal error: %s", err.Error())
	//	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
	//	return
	//}
	//
	//a.log.Infof("user weight: get user weight: success")
	//ctx.JSON(http.StatusOK, models.NewUserWeightResponse(uw))
}

// getExercises получает список упражнений
// @Summary Получение списка упражнений
// @Description Получает список упражнений с возможностью фильтрации
// @Tags Exercises
// @Security BearerAuth
// @Produce json
// @Param level_preparation query string false "Уровень подготовки (beginner, medium, sportsman)"
// @Param type_exercise query string false "Тип упражнения (cardio, upper_body, lower_body, full_body, flexibility)"
// @Param place_exercise query string false "Место выполнения (home, gym, street)"
// @Success 200 {array} models.ExerciseResponseModel "Список упражнений"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID пользователя"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /exercises [get]
func (a *API) getExercises(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("exercise error: get exercises: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("exercise error: get exercises history: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("exercise error: get exercises: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	le, err := a.CRUDService.ListExercise(ctx, userID, dto.ExerciseFilter{}, false)
	if err != nil {
		a.log.Errorf("exercise error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"exercise error: internal error": err.Error()})
		return
	}

	a.log.Infof("exercise: get exercises: success")
	ctx.JSON(http.StatusOK, models.NewExerciseResponseList(le))
}

// updateExercise обновляет упражнение
// @Summary Обновление упражнения
// @Description Обновляет существующее упражнение по ID
// @Tags Exercises
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uuid path string true "ID упражнения"
// @Param request body  models.ExerciseResponseModel true "Данные для обновления упражнения"
// @Success 200 {object} models.SuccessResponse "Упражнение успешно обновлено"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 404 {object} models.ErrorResponse "Упражнение не найдено"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /exercises/{uuid} [patch]
func (a *API) updateExercise(ctx *gin.Context) {

}

// deleteExercise удаляет упражнение
// @Summary Удаление упражнения
// @Description Удаляет упражнение по ID
// @Tags Exercises
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID упражнения"
// @Success 200 {object} models.SuccessResponse "Упражнение успешно удалено"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 404 {object} models.ErrorResponse "Упражнение не найдено"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /exercises/{uuid} [delete]
func (a *API) deleteExercise(ctx *gin.Context) {
	id := ctx.Param("uuid")
	if id == "" {
		a.log.Errorf("exercise error: create exercise: missing exercise id in header")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing exercise id in header",
		})
		return
	}

	exerciseID, err := uuid.Parse(id)
	if err != nil {
		a.log.Errorf("exercise error: delete exercise: invalid exercise id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid id format",
			"details": err.Error(),
		})
		return
	}

	err = a.CRUDService.DeleteExercise(ctx, dto.ExerciseFilter{ID: &exerciseID})
	if err != nil {
		a.log.Errorf("exercise error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"exercise error: internal error": err.Error()})
		return
	}

	a.log.Infof("exercise: delete exercise: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}

// createExercise создает новое упражнение
// @Summary Создание нового упражнения
// @Description Создает новое упражнение с указанными параметрами
// @Tags Exercises
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.ExerciseRequestModel true "Данные для создания упражнения"
// @Success 200 {object} models.SuccessResponse "Упражнение успешно создано"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации или неверный формат данных"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /exercises [post]
func (a *API) createExercise(ctx *gin.Context) {
	var m models.ExerciseRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("exercise error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"exercise error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "create exercise")
		return
	}
	e, err := m.ToSpec()
	if err != nil {
		a.log.Errorf("exercise error: create exercise: invalid data: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"exercise error: create exercise: invalid data": err.Error()})
		return
	}

	if err := a.CRUDService.CreateExercise(ctx, e); err != nil {
		a.log.Errorf("exercise error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"exercise error: internal error": err.Error()})
		return
	}

	a.log.Infof("exercise: create exercise: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully created"})
}
