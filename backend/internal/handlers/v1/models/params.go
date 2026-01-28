package models

import (
	"backend/internal/domain/entities"
	"fmt"
	"github.com/google/uuid"
)

type UserParamsResponseModel struct {
	Height              int                `json:"height"`
	Photo               string             `json:"photo"`
	Wants               entities.Want      `json:"wants"`
	Lifestyle           entities.Lifestyle `json:"lifestyle"`
	CurrentWeight       float64            `json:"currentWeight"`
	TargetWeight        float64            `json:"targetWeight"`
	TargetCaloriesDaily int                `json:"targetCaloriesDaily"`
	TargetWorkoutsWeeks int                `json:"targetWorkoutsWeeks"`
}

func NewUserParamsResponse(params *entities.UserParams) UserParamsResponseModel {
	return UserParamsResponseModel{Height: params.Height(),
		Photo:               params.Photo(),
		Wants:               params.Want(),
		Lifestyle:           params.Lifestyle(),
		CurrentWeight:       params.CurrentWeight(),
		TargetWeight:        params.TargetWeight(),
		TargetCaloriesDaily: params.TargetCaloriesDaily(),
		TargetWorkoutsWeeks: params.TargetWorkoutsWeeks(),
	}
}

type UserParamsUpdateRequestModel struct {
	Height              *int     `json:"height" validate:"omitempty,min=100,max=250"`
	Photo               *string  `json:"photo"`
	Wants               *string  `json:"wants" validate:"omitempty,oneof=lose_weight build_muscle stay_fit"`
	TargetWorkoutsWeeks *int     `json:"targetWorkoutsWeeks" validate:"omitempty,min=0,max=7"`
	TargetCaloriesDaily *int     `json:"targetCaloriesDaily" validate:"omitempty,min=0,max=10000"`
	TargetWeight        *float64 `json:"targetWeight" validate:"omitempty,min=40,max=300"`
	Lifestyle           *string  `json:"lifestyle" validate:"omitempty,oneof=not_active active sportive"`
}

func (up *UserParamsUpdateRequestModel) ToParam() (entities.UserParamsUpdateParams, error) {
	var wants *entities.Want
	if up.Wants != nil {
		w, err := entities.ToWant(*up.Wants)
		if err != nil {
			return entities.UserParamsUpdateParams{}, fmt.Errorf("invalid field want : %w", err)
		}
		wants = &w
	}

	var lifestyle *entities.Lifestyle
	if up.Lifestyle != nil {
		l, err := entities.ToLifestyle(*up.Lifestyle)
		if err != nil {
			return entities.UserParamsUpdateParams{}, fmt.Errorf("invalid field lifestyle : %w", err)
		}
		lifestyle = &l
	}

	return entities.UserParamsUpdateParams{
		Height:              up.Height,
		Photo:               up.Photo,
		Wants:               wants,
		Lifestyle:           lifestyle,
		TargetCaloriesDaily: up.TargetCaloriesDaily,
		TargetWorkoutsWeeks: up.TargetWorkoutsWeeks,
		TargetWeight:        up.TargetWeight,
	}, nil
}

type UserParamsCreateRequestModel struct {
	Height              *int     `json:"height" validate:"required,min=100,max=250"`
	Photo               *string  `json:"photo"`
	Wants               *string  `json:"wants" validate:"required,oneof=lose_weight build_muscle stay_fit"`
	Lifestyle           *string  `json:"lifestyle" validate:"required,oneof=not_active active sportive"`
	TargetCaloriesDaily *int     `json:"targetCaloriesDaily" validate:"required,min=0,max=10000"`
	TargetWorkoutsWeeks *int     `json:"targetWorkoutsWeeks" validate:"required,min=0,max=7"`
	TargetWeight        *float64 `json:"targetWeight" validate:"required,min=40,max=300"`
}

func (up *UserParamsCreateRequestModel) ToSpec() (entities.UserParamsInitSpec, error) {
	want, err := entities.ToWant(*up.Wants)
	if err != nil {
		return entities.UserParamsInitSpec{}, fmt.Errorf("invalid field want : %w", err)
	}
	lifestyle, err := entities.ToLifestyle(*up.Lifestyle)
	if err != nil {
		return entities.UserParamsInitSpec{}, fmt.Errorf("invalid field lifestyle : %w", err)
	}
	var height, targetCaloriesDaily, targetWorkoutsWeeks int
	var photo string
	var targetWeight float64
	if up.Height != nil {
		height = *up.Height
	}

	if up.Photo != nil {
		photo = *up.Photo
	}

	if up.TargetCaloriesDaily != nil {
		targetCaloriesDaily = *up.TargetCaloriesDaily
	}

	if up.TargetWorkoutsWeeks != nil {
		targetWorkoutsWeeks = *up.TargetWorkoutsWeeks
	}

	if up.TargetWeight != nil {
		targetWeight = *up.TargetWeight
	}

	return entities.UserParamsInitSpec{
		ID:                  uuid.New(),
		Height:              height,
		Photo:               photo,
		Wants:               want,
		Lifestyle:           lifestyle,
		TargetWorkoutsWeeks: targetWorkoutsWeeks,
		TargetWeight:        targetWeight,
		TargetCaloriesDaily: targetCaloriesDaily,
	}, nil
}
