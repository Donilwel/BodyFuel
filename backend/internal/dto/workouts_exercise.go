package dto

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type WorkoutsExerciseFilter struct {
	WorkoutID       *uuid.UUID
	ExerciseID      *uuid.UUID
	ModifyReps      *int
	ModifyRelaxTime *int
	Calories        *int
	Status          *entities.ExerciseStatus
	UpdateAt        *time.Time
}

// SkippedExerciseInfo aggregates skip history for one exercise across all workouts of a user.
type SkippedExerciseInfo struct {
	ExerciseID    uuid.UUID
	SkipCount     int
	LastSkippedAt time.Time
}

// ExerciseProgressInfo aggregates completion history for one exercise to drive progressive overload.
type ExerciseProgressInfo struct {
	ExerciseID     uuid.UUID
	TypeExercise   entities.ExerciseType
	PlaceExercise  entities.PlaceExercise
	LastReps       int // reps used in the most recent completed set
	LastRelaxTime  int // rest time used in the most recent completed set (seconds)
	CompletedCount int // total completions within the lookback window
	SkippedCount   int // total skips within the lookback window
}
