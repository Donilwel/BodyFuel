package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ()

func (a *API) registerWorkoutsHandlers(router *gin.RouterGroup) {
	workout := router.Group("/workouts")
	workout.GET("/:uuid", a.getUserWorkout)
	workout.DELETE("/:uuid", a.deleteUserWorkout)
	workout.PATCH("/:uuid", a.updateUserWorkout)
	workout.POST("", a.generateWorkout)
	workout.GET("/history", a.getUserWorkouts)
}

// getUserWorkout получает тренировку пользователя по ID
// @Summary Получение тренировки пользователя
// @Description Получает детальную информацию о тренировке пользователя по ID, включая список упражнений
// @Tags Workouts
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID тренировки"
// @Success 200 {object} models.WorkoutResponse "Детальная информация о тренировке"
// @Failure 400 {object} models.ErrorResponse "Неверный формат UUID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 404 {object} models.ErrorResponse "Тренировка не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /workouts/{uuid} [get]
func (a *API) getUserWorkout(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("user params error: get user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("user params error: get user params: invalid user_id type in context")
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

	workoutID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		a.log.Errorf("get user workout error: invalid workout UUID: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid workout UUID",
		})
		return
	}

	workoutFilter := dto.WorkoutsFilter{
		ID:     &workoutID,
		UserID: &userID,
	}

	workout, err := a.CRUDService.GetWorkout(ctx, workoutFilter, false)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "workout not found",
			})
			return
		}
		a.log.Errorf("get user workout error: failed to get workout: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get workout",
		})
		return
	}

	exercisesFilter := dto.WorkoutsExerciseFilter{
		WorkoutID: &workoutID,
	}

	workoutExercises, err := a.CRUDService.ListWorkoutsExercise(ctx, exercisesFilter)
	if err != nil {
		a.log.Errorf("get user workout error: failed to get workout exercises: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get workout exercises",
		})
		return
	}

	exerciseIDs := make([]uuid.UUID, len(workoutExercises))
	for i, we := range workoutExercises {
		exerciseIDs[i] = we.ExerciseID()
	}

	exercises, err := a.CRUDService.ListExercise(ctx, userID, dto.ExerciseFilter{}, false)
	if err != nil {
		a.log.Errorf("get user workout error: failed to get exercises details: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get exercises details",
		})
		return
	}

	exercisesMap := make(map[uuid.UUID]*entities.Exercise)
	for _, e := range exercises {
		exercisesMap[e.ID()] = e
	}

	response := models.WorkoutResponse{
		ID:                 workout.ID(),
		UserID:             workout.UserID(),
		Level:              workout.Level(),
		PredictionCalories: workout.PredictionCalories(),
		TotalCalories:      workout.TotalCalories(),
		Status:             workout.Status(),
		Duration:           workout.Duration(),
		CreatedAt:          workout.CreatedAt(),
		UpdatedAt:          workout.UpdatedAt(),
		Exercises:          make([]models.WorkoutExerciseResponse, 0, len(workoutExercises)),
	}

	for _, we := range workoutExercises {
		exercise := exercisesMap[we.ExerciseID()]
		if exercise == nil {
			a.log.Warnf("get user workout warning: exercise %s not found", we.ExerciseID())
			continue
		}

		var completedAt *time.Time
		if we.Status() == entities.ExerciseStatusCompleted {
			t := we.UpdatedAt()
			completedAt = &t
		}

		response.Exercises = append(response.Exercises, models.WorkoutExerciseResponse{
			ExerciseID:       we.ExerciseID(),
			Name:             exercise.Name(),
			Description:      exercise.Description(),
			TypeExercise:     exercise.TypeExercise(),
			PlaceExercise:    exercise.PlaceExercise(),
			LevelPreparation: exercise.LevelPreparation(),
			LinkGif:          exercise.LinkGif(),
			ModifyReps:       we.ModifyReps(),
			ModifyRelaxTime:  we.ModifyRelaxTime(),
			Status:           we.Status(),
			AvgCaloriesPer:   exercise.AvgCaloriesPer(),
			Steps:            exercise.Steps(),
			CompletedAt:      completedAt,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

// getUserWorkouts получает историю тренировок пользователя
// @Summary Получение истории тренировок пользователя
// @Description Получает список всех тренировок пользователя (без детализации упражнений)
// @Tags Workouts
// @Security BearerAuth
// @Produce json
// @Success 200 {array} []models.UserWorkoutResponse "История тренировок пользователя"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /workouts/history [get]
func (a *API) getUserWorkouts(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("user params error: get user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("user params error: get user params: invalid user_id type in context")
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

	workoutFilter := dto.WorkoutsFilter{
		UserID: &userID,
	}

	workouts, err := a.CRUDService.ListWorkouts(ctx, workoutFilter, false)
	if err != nil {
		a.log.Errorf("list workouts error: failed to get workouts: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get workout",
		})
		return
	}
	fmt.Println(workouts)
	ctx.JSON(http.StatusOK, models.NewUserWorkoutsResponse(workouts))
}

// updateUserWorkout обновляет тренировку пользователя
// @Summary Обновление тренировки пользователя
// @Description Обновляет информацию о тренировке пользователя (например, статус, длительность, калории)
// @Tags Workouts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uuid path string true "ID тренировки"
// @Success 200 {object} models.SuccessResponse "Успешное обновление"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации или неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 404 {object} models.ErrorResponse "Тренировка не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /workouts/{uuid} [patch]
func (a *API) updateUserWorkout(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("update workout error: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user_id in context"})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("update workout error: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("update workout error: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format", "details": err.Error()})
		return
	}

	workoutID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		a.log.Errorf("update workout error: invalid workout UUID: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid workout UUID"})
		return
	}

	var req models.UpdateWorkoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.log.Errorf("update workout error: invalid request body: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	now := time.Now()
	params := entities.WorkoutUpdateParams{
		Status:    req.Status,
		UpdatedAt: &now,
	}
	if req.Duration != nil {
		d := req.Duration.Nanoseconds()
		params.Duration = &d
	}

	f := dto.WorkoutsFilter{
		ID:     &workoutID,
		UserID: &userID,
	}

	if err := a.CRUDService.UpdateWorkoutByFilter(ctx, f, params); err != nil {
		a.log.Errorf("update workout error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update workout", "details": err.Error()})
		return
	}

	a.log.Infof("update workout %s: success", workoutID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully updated"})
}

// deleteUserWorkout удаляет тренировку пользователя
// @Summary Удаление тренировки пользователя
// @Description Удаляет тренировку пользователя по ID
// @Tags Workouts
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID тренировки"
// @Success 200 {object} models.SuccessResponse "Успешное удаление"
// @Failure 400 {object} models.ErrorResponse "Неверный формат UUID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 404 {object} models.ErrorResponse "Тренировка не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /workouts/{uuid} [delete]
func (a *API) deleteUserWorkout(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("user params error: get user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("user params error: get user params: invalid user_id type in context")
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

	workoutID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		a.log.Errorf("get user workout error: invalid workout UUID: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid workout UUID",
		})
		return
	}

	workoutFilter := dto.WorkoutsFilter{
		ID:     &workoutID,
		UserID: &userID,
	}

	if err := a.CRUDService.DeleteWorkout(ctx, workoutFilter); err != nil {
		a.log.Errorf("user workout error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"user workout error: internal error": err.Error()})
		return
	}

	a.log.Infof("user workouts: delete user workout: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})

}

// generateWorkout генерирует тренировку по параметрам пользователя
// @Summary Генерация тренировки по параметрам
// @Description Генерирует тренировку на основе указанных параметров (место, тип, уровень, количество упражнений)
// @Tags Workouts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.GenerateWorkoutRequest true "Параметры генерации тренировки"
// @Success 201 {object} models.WorkoutResponse "Созданная тренировка"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /workouts [post]
func (a *API) generateWorkout(ctx *gin.Context) {
	// Получаем user_id из контекста
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("user params error: get user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("user params error: get user params: invalid user_id type in context")
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

	var req models.GenerateWorkoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.log.Errorf("generate workout error: invalid request format: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		a.log.Errorf("generate workout error: validation failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	userParams, err := a.CRUDService.GetParamsUser(ctx, dto.UserParamsFilter{UserID: &userID}, false)
	if err != nil {
		if err == sql.ErrNoRows {
			a.log.Warnf("generate workout: no user params found for user %s, using defaults", userID)
		} else {
			a.log.Errorf("generate workout error: failed to get user params: %v", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get user params",
			})
			return
		}
	}

	userInfo, err := a.CRUDService.GetInfoUser(ctx, dto.UserInfoFilter{ID: &userID}, false)
	if err != nil && err != sql.ErrNoRows {
		a.log.Errorf("generate workout error: failed to get user info: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get user info",
		})
		return
	}

	generateParams := &dto.GenerateWorkoutParams{
		UserID:         userID,
		UserParams:     userParams,
		UserInfo:       userInfo,
		PlaceExercise:  req.PlaceExercise,
		TypeExercise:   req.TypeExercise,
		Level:          req.Level,
		ExercisesCount: req.ExercisesCount,
	}

	workout, err := a.WorkoutService.GenerateCustomWorkout(ctx, generateParams)
	if err != nil {
		a.log.Errorf("generate workout error: failed to generate workout: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate workout",
		})
		return
	}

	workoutId := workout.ID()
	exercisesFilter := dto.WorkoutsExerciseFilter{
		WorkoutID: &workoutId,
	}

	workoutExercises, err := a.CRUDService.ListWorkoutsExercise(ctx, exercisesFilter)
	if err != nil {
		a.log.Errorf("generate workout error: failed to get workout exercises: %v", err)
		ctx.JSON(http.StatusCreated, models.WorkoutResponse{
			ID:                 workout.ID(),
			UserID:             workout.UserID(),
			Level:              workout.Level(),
			PredictionCalories: workout.PredictionCalories(),
			TotalCalories:      workout.TotalCalories(),
			Status:             workout.Status(),
			Duration:           workout.Duration(),
			CreatedAt:          workout.CreatedAt(),
			UpdatedAt:          workout.UpdatedAt(),
			Exercises:          []models.WorkoutExerciseResponse{},
		})
		return
	}

	exerciseIDs := make([]uuid.UUID, len(workoutExercises))
	for i, we := range workoutExercises {
		exerciseIDs[i] = we.ExerciseID()
	}

	exercises, err := a.CRUDService.ListExercise(ctx, userID, dto.ExerciseFilter{}, false)
	if err != nil {
		a.log.Errorf("generate workout error: failed to get exercises details: %v", err)
	}

	exercisesMap := make(map[uuid.UUID]*entities.Exercise)
	for _, e := range exercises {
		exercisesMap[e.ID()] = e
	}

	response := models.WorkoutResponse{
		ID:                 workout.ID(),
		UserID:             workout.UserID(),
		Level:              workout.Level(),
		PredictionCalories: workout.PredictionCalories(),
		TotalCalories:      workout.TotalCalories(),
		Status:             workout.Status(),
		Duration:           workout.Duration(),
		CreatedAt:          workout.CreatedAt(),
		UpdatedAt:          workout.UpdatedAt(),
		Exercises:          make([]models.WorkoutExerciseResponse, 0, len(workoutExercises)),
	}

	for _, we := range workoutExercises {
		exercise := exercisesMap[we.ExerciseID()]
		if exercise == nil {
			a.log.Warnf("generate workout warning: exercise %s not found", we.ExerciseID())
			continue
		}

		var completedAt *time.Time
		if we.Status() == entities.ExerciseStatusCompleted {
			t := we.UpdatedAt()
			completedAt = &t
		}

		response.Exercises = append(response.Exercises, models.WorkoutExerciseResponse{
			ExerciseID:       we.ExerciseID(),
			Name:             exercise.Name(),
			Description:      exercise.Description(),
			TypeExercise:     exercise.TypeExercise(),
			PlaceExercise:    exercise.PlaceExercise(),
			LevelPreparation: exercise.LevelPreparation(),
			LinkGif:          exercise.LinkGif(),
			ModifyReps:       we.ModifyReps(),
			ModifyRelaxTime:  we.ModifyRelaxTime(),
			Status:           we.Status(),
			AvgCaloriesPer:   exercise.AvgCaloriesPer(),
			Steps:            exercise.Steps(),
			CompletedAt:      completedAt,
		})
	}

	a.log.Infof("Successfully generated custom workout %s for user %s", workout.ID(), userID)
	ctx.JSON(http.StatusCreated, response)
}
