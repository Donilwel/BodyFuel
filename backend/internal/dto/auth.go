package dto

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
)

type UserRefreshTokenFilter struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	TokenHash *string
}

type UserVerificationCodeFilter struct {
	UserID   *uuid.UUID
	CodeType *entities.VerificationCodeType
}
