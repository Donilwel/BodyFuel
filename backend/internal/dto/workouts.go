package dto

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type WorkoutsFilter struct {
	ID                 *uuid.UUID
	UserID             *uuid.UUID
	Level              *entities.WorkoutsLevel
	TotalCalories      *int
	PredictionCalories *int
	Status             *entities.WorkoutsStatus
	Duration           *time.Duration
	CreatedAt          *time.Time
	UpdatedAt          *time.Time
}

type WorkoutsExercise struct {
	WorkoutID       *uuid.UUID
	ExerciseID      *uuid.UUID
	ModifyReps      int
	ModifyRelaxTime int
	Calories        int
	Status          *entities.ExerciseStatus
	UpdateAt        *time.Time
}
