package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type WorkoutRow struct {
	ID                 uuid.UUID               `db:"id"`
	UserID             uuid.UUID               `db:"user_id"`
	Level              entities.WorkoutsLevel  `db:"level"`
	Status             entities.WorkoutsStatus `db:"status"`
	TotalCalories      int                     `db:"total_calories"`
	PredictionCalories int                     `db:"prediction_calories"`
	Duration           time.Duration           `db:"duration"`
	CreatedAt          time.Time               `db:"created_at"`
	UpdatedAt          time.Time               `db:"updated_at"`
}

type WorkoutsExercise struct {
	WorkoutID  uuid.UUID
	ExerciseID uuid.UUID
	Status     entities.ExerciseStatus
}

func NewWorkoutRow(workout *entities.Workout) *WorkoutRow {
	return &WorkoutRow{
		ID:                 workout.ID(),
		UserID:             workout.UserID(),
		Level:              workout.Level(),
		TotalCalories:      workout.TotalCalories(),
		PredictionCalories: workout.PredictionCalories(),
		Status:             workout.Status(),
		Duration:           workout.Duration(),
		CreatedAt:          workout.CreatedAt(),
		UpdatedAt:          workout.UpdatedAt(),
	}
}

func (u *WorkoutRow) ToEntity() *entities.Workout {
	return entities.NewWorkout(
		entities.WithWorkoutRestoreSpec(entities.WorkoutRestoreSpec{
			ID:                 u.ID,
			UserID:             u.UserID,
			Level:              u.Level,
			TotalCalories:      u.TotalCalories,
			PredictionCalories: u.PredictionCalories,
			Status:             u.Status,
			Duration:           u.Duration,
			CreatedAt:          u.CreatedAt,
			UpdatedAt:          u.UpdatedAt,
		}),
	)
}
