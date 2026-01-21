package models

import (
	"backend/internal/domain/entities"
	"fmt"
	"github.com/google/uuid"
)

type UserParamsRequestModel struct {
}

func (r *UserParamsRequestModel) ToSpec(userID uuid.UUID) entities.UserParamsInitSpec {
	return entities.UserParamsInitSpec{
		ID: uuid.New(),
	}
}

type UserParamsResponseModel struct {
	Height    int                `json:"height"`
	Photo     string             `json:"photo"`
	Wants     entities.Want      `json:"wants"`
	Lifestyle entities.Lifestyle `json:"lifestyle"`
}

func NewUserParamsResponse(params *entities.UserParams) UserParamsResponseModel {
	return UserParamsResponseModel{Height: params.Height(),
		Photo:     params.Photo(),
		Wants:     params.Want(),
		Lifestyle: params.Lifestyle(),
	}
}

type UserParamsUpdateRequestModel struct {
	Height    *int    `json:"height" validate:"omitempty,min=100,max=250"`
	Photo     *string `json:"photo"`
	Wants     *string `json:"wants" validate:"oneof=lose_weight build_muscle stay_fit"`
	Lifestyle *string `json:"lifestyle" validate:"oneof=not_active active sportive"`
}

func (up *UserParamsUpdateRequestModel) ToParam() (entities.UserParamsUpdateParams, error) {
	want, err := entities.ToWant(*up.Wants)
	if err != nil {
		return entities.UserParamsUpdateParams{}, fmt.Errorf("invalid field want : %w", err)
	}
	lifestyle, err := entities.ToLifestyle(*up.Lifestyle)
	if err != nil {
		return entities.UserParamsUpdateParams{}, fmt.Errorf("invalid field lifestyle : %w", err)
	}
	return entities.UserParamsUpdateParams{
		Height:    up.Height,
		Photo:     up.Photo,
		Wants:     &want,
		Lifestyle: &lifestyle,
	}, nil
}

type UserParamsCreateRequestModel struct {
	Height    *int    `json:"height" validate:"required,min=100,max=250"`
	Photo     *string `json:"photo"`
	Wants     *string `json:"wants" validate:"required,oneof=lose_weight build_muscle stay_fit"`
	Lifestyle *string `json:"lifestyle" validate:"required,oneof=not_active active sportive"`
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
	var height int
	var photo string
	if up.Height != nil {
		height = *up.Height
	}

	if up.Photo != nil {
		photo = *up.Photo
	}
	return entities.UserParamsInitSpec{
		ID:        uuid.New(),
		Height:    height,
		Photo:     photo,
		Wants:     want,
		Lifestyle: lifestyle,
	}, nil
}
