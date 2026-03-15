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
