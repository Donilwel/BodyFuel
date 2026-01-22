package dto

import (
	"github.com/google/uuid"
	"time"
)

type UserWeightFilter struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	Weight    *float64
	CreatedAt *time.Time
}

type UserWeight struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	Weight    *float64
	CreatedAt *string
}
