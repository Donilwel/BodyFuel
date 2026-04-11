package dto

import (
	"github.com/google/uuid"
	"time"
)

type UserCaloriesFilter struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
}
