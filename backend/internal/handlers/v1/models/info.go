package models

import (
	"backend/internal/domain/entities"
	"time"
)

type UserInfoResponseModel struct {
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserInfoResponse(params *entities.UserInfo) UserInfoResponseModel {
	return UserInfoResponseModel{
		params.Username(),
		params.Name(),
		params.Surname(),
		params.Email(),
		params.Phone(),
		params.CreatedAt(),
	}
}

type UserInfoUpdateRequestModel struct {
	Name    *string `json:"name" form:"name" validate:"required,min=2,max=50"`
	Surname *string `json:"surname" form:"surname" validate:"required,min=2,max=50"`
	Email   *string `json:"email" form:"email" validate:"required"`
	Phone   *string `json:"phone" form:"phone" validate:"required,regex=^\\+?[0-9]{10,15}$"`
}

func (u *UserInfoUpdateRequestModel) ToParam() entities.UserInfoUpdateParams {
	return entities.UserInfoUpdateParams{
		Name:    u.Name,
		Surname: u.Surname,
		Email:   u.Email,
		Phone:   u.Phone,
	}
}
