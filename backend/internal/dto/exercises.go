package dto

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
)

type ExerciseFilter struct {
	ID               *uuid.UUID
	LevelPreparation *entities.LevelPreparation
	Name             *string
	TypeExercise     *entities.ExerciseType
	Description      *string
	BaseCountReps    *int
	Steps            *int
	LinkGif          *string
	PlaceExercise    *entities.PlaceExercise
	AvgCaloriesPer   *float64
	BaseRelaxTime    *int
}
