package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type UserInfo struct {
	ID        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	Name      string    `db:"name"`
	Surname   string    `db:"surname"`
	Password  string    `db:"password"`
	Email     string    `db:"email"`
	Phone     string    `db:"phone"`
	CreatedAt time.Time `db:"created_at"`
}

func NewUserInfoRow(userInfo *entities.UserInfo) *UserInfo {
	return &UserInfo{
		ID:        userInfo.ID(),
		Username:  userInfo.Username(),
		Name:      userInfo.Name(),
		Surname:   userInfo.Surname(),
		Password:  userInfo.Password(),
		Email:     userInfo.Email(),
		Phone:     userInfo.Phone(),
		CreatedAt: userInfo.CreatedAt(),
	}
}

func (u *UserInfo) ToEntity() *entities.UserInfo {
	return entities.NewUserInfo(
		entities.WithUserInfoRestoreSpec(entities.UserInfoRestoreSpec{
			ID:        u.ID,
			Username:  u.Username,
			Name:      u.Name,
			Surname:   u.Surname,
			Password:  u.Password,
			Email:     u.Email,
			Phone:     u.Phone,
			CreatedAt: u.CreatedAt,
		}),
	)
}
