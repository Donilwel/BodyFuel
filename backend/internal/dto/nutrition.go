package dto

import (
	"github.com/google/uuid"
	"time"
)

type UserFoodFilter struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	MealType  *string
	Date      *time.Time
	StartDate *time.Time
	EndDate   *time.Time
}

type UserRecommendationFilter struct {
	ID     *uuid.UUID
	UserID *uuid.UUID
	IsRead *bool
	Limit  *int
	Offset *int
}
