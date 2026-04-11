package models

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type CreateUserCaloriesRequest struct {
	Calories    int       `json:"calories" validate:"required,min=0,max=10000"`
	Date        time.Time `json:"date" validate:"required"`
	Description string    `json:"description" validate:"omitempty,max=255"`
}

type UpdateUserCaloriesRequest struct {
	Calories    *int       `json:"calories" validate:"omitempty,min=0,max=10000"`
	Date        *time.Time `json:"date" validate:"omitempty"`
	Description *string    `json:"description" validate:"omitempty,max=255"`
}

type UserCaloriesResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Calories    int       `json:"calories"`
	Description string    `json:"description,omitempty"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewUserCaloriesResponse(uc *entities.UserCalories) UserCaloriesResponse {
	return UserCaloriesResponse{
		ID:          uc.ID(),
		UserID:      uc.UserID(),
		Calories:    uc.Calories(),
		Description: uc.Description(),
		Date:        uc.Date(),
		CreatedAt:   uc.CreatedAt(),
		UpdatedAt:   uc.UpdatedAt(),
	}
}

func NewUserCaloriesResponseList(list []*entities.UserCalories) []UserCaloriesResponse {
	resp := make([]UserCaloriesResponse, len(list))
	for i, uc := range list {
		resp[i] = NewUserCaloriesResponse(uc)
	}
	return resp
}
