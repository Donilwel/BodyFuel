package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
)

type UserParams struct {
	ID        uuid.UUID          `db:"id"`
	UserId    uuid.UUID          `db:"id_user"`
	Height    int                `db:"height"`
	Photo     string             `db:"photo"`
	Wants     entities.Want      `db:"wants"`
	Lifestyle entities.Lifestyle `db:"lifestyle"`
}

func NewUserParamsRow(userParams *entities.UserParams) *UserParams {
	return &UserParams{
		ID:        userParams.ID(),
		UserId:    userParams.UserID(),
		Height:    userParams.Height(),
		Photo:     userParams.Photo(),
		Wants:     userParams.Want(),
		Lifestyle: userParams.Lifestyle(),
	}
}

func (u *UserParams) ToEntity() *entities.UserParams {
	return entities.NewUserParams(
		entities.WithUserParamsRestoreSpec(entities.UserParamsRestoreSpec{
			ID:        u.ID,
			UserID:    u.UserId,
			Height:    u.Height,
			Photo:     u.Photo,
			Wants:     u.Wants,
			Lifestyle: u.Lifestyle,
		}),
	)
}
