package models

import (
	"backend/internal/domain/entities"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UpdateWorkoutRequest содержит поля для обновления тренировки.
// Duration задаётся в наносекундах (int64).
type UpdateWorkoutRequest struct {
	Status   *entities.WorkoutsStatus `json:"status"   binding:"omitempty,oneof=pending in_progress completed cancelled"`
	Duration *int64                   `json:"duration" binding:"omitempty,min=60000000000"`
}

type WorkoutResponse struct {
	ID                 uuid.UUID                 `json:"id"`
	UserID             uuid.UUID                 `json:"user_id"`
	Level              entities.WorkoutsLevel    `json:"level"`
	TotalCalories      int                       `json:"total_calories"`
	PredictionCalories int                       `json:"prediction_calories"`
	Status             entities.WorkoutsStatus   `json:"status"`
	Duration           int64                     `json:"duration,omitempty"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
	Exercises          []WorkoutExerciseResponse `json:"exercises,omitempty"`
}

type WorkoutExerciseResponse struct {
	ExerciseID       uuid.UUID                 `json:"exercise_id"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	TypeExercise     entities.ExerciseType     `json:"type_exercise"`
	PlaceExercise    entities.PlaceExercise    `json:"place_exercise"`
	LevelPreparation entities.LevelPreparation `json:"level_preparation"`
	LinkGif          string                    `json:"link_gif"`
	ModifyReps       int                       `json:"modify_reps"`
	ModifyRelaxTime  int                       `json:"modify_relax_time"`
	Status           entities.ExerciseStatus   `json:"status"`
	AvgCaloriesPer   float64                   `json:"avg_calories_per"`
	Steps            int                       `json:"steps"`
	CompletedAt      *time.Time                `json:"completed_at,omitempty"`
}

type WorkoutHistoryResponse struct {
	Workouts []WorkoutSummaryResponse `json:"workouts"`
	Total    int                      `json:"total"`
	Limit    int                      `json:"limit"`
	Offset   int                      `json:"offset"`
}

type WorkoutSummaryResponse struct {
	ID             uuid.UUID               `json:"id"`
	Level          entities.WorkoutsLevel  `json:"level"`
	TotalCalories  int                     `json:"total_calories"`
	Status         entities.WorkoutsStatus `json:"status"`
	Duration       *time.Duration          `json:"duration,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	ExercisesCount int                     `json:"exercises_count"`
	CompletedCount int                     `json:"completed_count"`
}

type WorkoutExercisesResponse struct {
	ExerciseID      uuid.UUID               `json:"exercise_id"`
	ModifyReps      int                     `json:"modify_reps"`
	ModifyRelaxTime int                     `json:"modify_relax_time"`
	Calories        int                     `json:"calories"`
	Status          entities.ExerciseStatus `json:"status"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

type UserWorkoutResponse struct {
	ID                 uuid.UUID               `json:"id"`
	UserID             uuid.UUID               `json:"user_id"`
	Level              entities.WorkoutsLevel  `json:"level"`
	Status             entities.WorkoutsStatus `json:"status"`
	Duration           int64                   `json:"duration,omitempty"`
	PredictionCalories int                     `json:"prediction_calories"`
	TotalCalories      int                     `json:"total_calories"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

func NewUserWorkoutResponse(w *entities.Workout) UserWorkoutResponse {
	return UserWorkoutResponse{
		ID:                 w.ID(),
		UserID:             w.UserID(),
		Level:              w.Level(),
		Status:             w.Status(),
		Duration:           w.Duration(),
		PredictionCalories: w.PredictionCalories(),
		TotalCalories:      w.TotalCalories(),
		CreatedAt:          w.CreatedAt(),
		UpdatedAt:          w.UpdatedAt(),
	}
}

func NewUserWorkoutsResponse(ws []*entities.Workout) []UserWorkoutResponse {
	var response []UserWorkoutResponse
	for _, w := range ws {
		response = append(response, NewUserWorkoutResponse(w))
	}
	return response
}

type AddWorkoutExerciseRequest struct {
	ExerciseID      uuid.UUID `json:"exercise_id" validate:"required"`
	ModifyReps      int       `json:"modify_reps" validate:"omitempty,min=1,max=1000"`
	ModifyRelaxTime int       `json:"modify_relax_time" validate:"omitempty,min=0,max=3600"`
}

type UpdateWorkoutExerciseRequest struct {
	Status          *string `json:"status" validate:"omitempty,oneof=pending in_progress completed skipped"`
	ModifyReps      *int    `json:"modify_reps" validate:"omitempty,min=1,max=1000"`
	ModifyRelaxTime *int    `json:"modify_relax_time" validate:"omitempty,min=0,max=3600"`
	Calories        *int    `json:"calories" validate:"omitempty,min=0"`
}

type WorkoutExerciseFullResponse struct {
	ExerciseID      uuid.UUID               `json:"exercise_id"`
	WorkoutID       uuid.UUID               `json:"workout_id"`
	ModifyReps      int                     `json:"modify_reps"`
	ModifyRelaxTime int                     `json:"modify_relax_time"`
	Calories        int                     `json:"calories"`
	Status          entities.ExerciseStatus `json:"status"`
	OrderIndex      int                     `json:"order_index"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

func NewWorkoutExerciseFullResponse(we *entities.WorkoutsExercise) WorkoutExerciseFullResponse {
	return WorkoutExerciseFullResponse{
		ExerciseID:      we.ExerciseID(),
		WorkoutID:       we.WorkoutID(),
		ModifyReps:      we.ModifyReps(),
		ModifyRelaxTime: we.ModifyRelaxTime(),
		Calories:        we.Calories(),
		Status:          we.Status(),
		OrderIndex:      we.OrderIndex(),
		CreatedAt:       we.CreatedAt(),
		UpdatedAt:       we.UpdatedAt(),
	}
}

func NewWorkoutExercisesFullResponse(list []*entities.WorkoutsExercise) []WorkoutExerciseFullResponse {
	resp := make([]WorkoutExerciseFullResponse, len(list))
	for i, we := range list {
		resp[i] = NewWorkoutExerciseFullResponse(we)
	}
	return resp
}

type GenerateWorkoutRequest struct {
	PlaceExercise  *entities.PlaceExercise `json:"place_exercise" binding:"omitempty,oneof=home gym street"`
	TypeExercise   *entities.ExerciseType  `json:"type_exercise" binding:"omitempty,oneof=upper_body lower_body full_body cardio flexibility"`
	Level          *entities.WorkoutsLevel `json:"level" binding:"omitempty,oneof=workout_light workout_middle workout_hard"`
	ExercisesCount *int                    `json:"exercises_count" binding:"omitempty,min=4,max=20"`
}

func (r *GenerateWorkoutRequest) Validate() error {
	if r.ExercisesCount != nil && (*r.ExercisesCount < 4 || *r.ExercisesCount > 20) {
		return fmt.Errorf("exercises_count must be between 4 and 20")
	}
	return nil
}
