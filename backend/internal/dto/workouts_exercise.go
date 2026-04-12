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
