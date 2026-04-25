package dto

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type WorkoutsFilter struct {
	ID                 *uuid.UUID
	UserID             *uuid.UUID
	Level              *entities.WorkoutsLevel
	TotalCalories      *int
	PredictionCalories *int
	Status             *entities.WorkoutsStatus
	Duration           *int64
	CreatedAt          *time.Time
	CreatedFrom        *time.Time
	CreatedTo          *time.Time
	UpdatedAt          *time.Time
}

type GenerateWorkoutParams struct {
	UserID         uuid.UUID
	UserParams     *entities.UserParams
	UserInfo       *entities.UserInfo
	PlaceExercise  *entities.PlaceExercise
	TypeExercise   *entities.ExerciseType
	Level          *entities.WorkoutsLevel
	ExercisesCount *int
}
