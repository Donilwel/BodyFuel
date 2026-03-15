package dto

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type WorkoutsFilter struct {
	ID                 *uuid.UUID
	UserID             *uuid.UUID
	Level              *entities.WorkoutsLevel
	TotalCalories      *int
	PredictionCalories *int
	Status             *entities.WorkoutsStatus
	Duration           *time.Duration
	CreatedAt          *time.Time
	UpdatedAt          *time.Time
}
