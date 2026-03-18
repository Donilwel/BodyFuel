package models

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type GenerateWorkoutRequest struct {
	Level    *entities.WorkoutsLevel `json:"level" binding:"required,oneof=beginner intermediate advanced"`
	Duration *time.Duration          `json:"duration" binding:"omitempty,min=60000000000"` // мин 1 минута
}

type UpdateWorkoutRequest struct {
	Status   *entities.WorkoutsStatus `json:"status" binding:"omitempty,oneof=pending in_progress completed cancelled"`
	Duration *time.Duration           `json:"duration" binding:"omitempty,min=60000000000"`
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
