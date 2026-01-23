package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type LoginRequestModel struct {
	Username string `json:"username" validate:"required" form:"username"`
	Password string `json:"password" validate:"required" form:"password"`
}

func (l *LoginRequestModel) ToSpec() entities.UserAuthInitSpec {
	return entities.UserAuthInitSpec{
		Username: l.Username,
		Password: l.Password,
	}
}

type RegisterRequestModel struct {
	Username string `json:"username" form:"username" validate:"required,min=3,max=32"`
	Name     string `json:"name" form:"name" validate:"required,min=2,max=50"`
	Surname  string `json:"surname" form:"surname" validate:"required,min=2,max=50"`
	Password string `json:"password,omitempty" form:"password" validate:"required,min=6"`
	Email    string `json:"email" form:"email" validate:"required"`
	Phone    string `json:"phone" form:"phone" validate:"required"`
}

func (r *RegisterRequestModel) ToSpec() entities.UserInfoInitSpec {
	return entities.UserInfoInitSpec{
		ID:        uuid.New(),
		Username:  r.Username,
		Name:      r.Name,
		Surname:   r.Surname,
		Email:     r.Email,
		Phone:     r.Phone,
		Password:  r.Password,
		CreatedAt: time.Now(),
	}
}
