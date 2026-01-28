package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
)

type ExerciseResponseModel struct {
	ID               uuid.UUID                 `json:"id"`
	LevelPreparation entities.LevelPreparation `json:"level_preparation"`
	Name             string                    `json:"name"`
	TypeExercise     entities.ExerciseType     `json:"type_exercise"`
	Description      string                    `json:"description"`
	BaseCountReps    int                       `json:"base_count_reps"`
	Steps            int                       `json:"steps"`
	LinkGif          string                    `json:"link_gif"`
	PlaceExercise    entities.PlaceExercise    `json:"place_exercise"`
	AvgCaloriesPer   float64                   `json:"avg_calories_per"`
	BaseRelaxTime    int                       `json:"base_relax_time"`
}

func NewExerciseResponse(params *entities.Exercise) ExerciseResponseModel {
	return ExerciseResponseModel{
		ID:               params.ID(),
		LevelPreparation: params.LevelPreparation(),
		Name:             params.Name(),
		TypeExercise:     params.TypeExercise(),
		Description:      params.Description(),
		BaseCountReps:    params.BaseCountReps(),
		Steps:            params.Steps(),
		LinkGif:          params.LinkGif(),
		PlaceExercise:    params.PlaceExercise(),
		AvgCaloriesPer:   params.AvgCaloriesPer(),
		BaseRelaxTime:    params.BaseRelaxTime(),
	}
}

func NewExerciseResponseList(weights []*entities.Exercise) []ExerciseResponseModel {
	var response []ExerciseResponseModel
	for _, weight := range weights {
		response = append(response, NewExerciseResponse(weight))
	}
	return response
}
