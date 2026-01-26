package dto

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
)

type UserParams struct {
	ID                  *uuid.UUID
	UserID              *uuid.UUID
	Height              *int
	Photo               *string
	Wants               *entities.Want
	TargetWorkoutsWeeks *int
	TargetCaloriesDaily *int
	TargetWeight        *float64
	Lifestyle           *entities.Lifestyle
}

type UserParamsFilter struct {
	ID                  *uuid.UUID
	UserID              *uuid.UUID
	Height              *int
	Photo               *string
	Wants               *entities.Want
	TargetWorkoutsWeeks *int
	TargetCaloriesDaily *int
	TargetWeight        *float64
	Lifestyle           *entities.Lifestyle
}
