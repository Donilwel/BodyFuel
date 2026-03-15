package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"database/sql"
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
		ID:            workout.ID(),
		UserID:        workout.UserID(),
		Level:         workout.Level(),
		TotalCalories: workout.TotalCalories(),
		Status:        workout.Status(),
		Duration:      workout.Duration(),
		CreatedAt:     workout.CreatedAt(),
		UpdatedAt:     workout.UpdatedAt(),
		Exercises:     make([]models.WorkoutExerciseResponse, 0, len(workoutExercises)),
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
			BaseCountReps:    exercise.BaseCountReps(),
			BaseRelaxTime:    exercise.BaseRelaxTime(),
			ModifyReps:       we.ModifyReps(),
			ModifyRelaxTime:  we.ModifyRelaxTime(),
			Calories:         we.Calories(),
			Status:           we.Status(),
			AvgCaloriesPer:   exercise.AvgCaloriesPer(),
			Steps:            exercise.Steps(),
			CompletedAt:      completedAt,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

func (a *API) getUserWorkouts(ctx *gin.Context) {

}

func (a *API) updateUserWorkout(ctx *gin.Context) {

}

func (a *API) deleteUserWorkout(ctx *gin.Context) {

}

func (a *API) generateWorkout(ctx *gin.Context) {

}
