package models

import (
	"backend/internal/domain/entities"
	"fmt"
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

type ExerciseRequestModel struct {
	LevelPreparation *string  `json:"level_preparation" validate:"required,oneof=beginner medium sportsman"`
	Name             *string  `json:"name" validate:"required,min=1,max=100"`
	TypeExercise     *string  `json:"type_exercise" validate:"required,oneof=cardio upper_body lower_body full_body flexibility"`
	Description      *string  `json:"description" validate:"omitempty,min=1,max=1000"`
	BaseCountReps    *int     `json:"base_count_reps" validate:"required,min=1,max=1000"`
	Steps            *int     `json:"steps" validate:"required,min=1,max=100"`
	LinkGif          *string  `json:"link_gif" validate:"omitempty,url"`
	PlaceExercise    *string  `json:"place_exercise" validate:"required,oneof=home gym street"`
	AvgCaloriesPer   *float64 `json:"avg_calories_per" validate:"required,min=0,max=1000"`
	BaseRelaxTime    *int     `json:"base_relax_time" validate:"required,min=0,max=3600"`
}

func (e *ExerciseRequestModel) ToSpec() (entities.ExerciseInitSpec, error) {
	var (
		linkGif     string
		description string
	)

	if e.LinkGif != nil {
		linkGif = *e.LinkGif
	}
	if e.Description != nil {
		description = *e.Description
	}

	levelPrep, err := entities.ToLevelPreparation(*e.LevelPreparation)
	if err != nil {
		return entities.ExerciseInitSpec{}, fmt.Errorf("invalid field level preparation: %w", err)
	}
	typeExercise, err := entities.ToExerciseType(*e.TypeExercise)
	if err != nil {
		return entities.ExerciseInitSpec{}, fmt.Errorf("invalid field exercise type: %w", err)
	}
	placeExercise, err := entities.ToPlaceExercise(*e.PlaceExercise)
	if err != nil {
		return entities.ExerciseInitSpec{}, fmt.Errorf("invalid field place exercise: %w", err)
	}

	return entities.ExerciseInitSpec{
		ID:               uuid.New(),
		LevelPreparation: levelPrep,
		Name:             *e.Name,
		TypeExercise:     typeExercise,
		Description:      description,
		BaseCountReps:    *e.BaseCountReps,
		Steps:            *e.Steps,
		LinkGif:          linkGif,
		PlaceExercise:    placeExercise,
		AvgCaloriesPer:   *e.AvgCaloriesPer,
		BaseRelaxTime:    *e.BaseRelaxTime,
	}, nil
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
