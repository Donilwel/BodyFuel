package dto

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
)

type UserParams struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	Height    *int
	Weight    []*entities.WeightDay
	Photo     *string
	Wants     []*entities.WantDay
	Lifestyle *entities.Lifestyle
}

type UserParamsFilter struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	Height    *int
	Weight    []*entities.WeightDay
	Photo     *string
	Wants     []*entities.WantDay
	Lifestyle *entities.Lifestyle
}
