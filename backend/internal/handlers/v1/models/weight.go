package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type UserWeightResponseModel struct {
	Weight float64   `json:"weight"`
	Date   time.Time `json:"date"`
}

func NewUserWeightResponse(weight *entities.UserWeight) UserWeightResponseModel {
	return UserWeightResponseModel{
		weight.Weight(),
		weight.Date(),
	}
}

func NewUserWeightResponseList(weights []*entities.UserWeight) []UserWeightResponseModel {
	var response []UserWeightResponseModel
	for _, weight := range weights {
		response = append(response, NewUserWeightResponse(weight))
	}
	return response
}

type UserWeightCreateRequestModel struct {
	Weight *float64 `json:"weight" validate:"required,min=10,max=300"`
}

func (u *UserWeightCreateRequestModel) ToSpec() entities.UserWeightInitSpec {
	var weight float64
	if u.Weight != nil {
		weight = *u.Weight
	}
	return entities.UserWeightInitSpec{
		ID:     uuid.New(),
		Weight: weight,
		Date:   time.Now(),
	}
}
