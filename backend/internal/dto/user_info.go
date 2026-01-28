package dto

import (
	"github.com/google/uuid"
	"time"
)

type UserInfoFilter struct {
	ID        *uuid.UUID
	Username  *string
	Name      *string
	Surname   *string
	Password  *string
	Email     *string
	Phone     *string
	CreatedAt *time.Time
}
