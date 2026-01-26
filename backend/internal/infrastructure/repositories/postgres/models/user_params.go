package models

import (
	"backend/internal/domain/entities"
	"database/sql"
	"github.com/google/uuid"
)

type UserParamsRow struct {
	ID                  uuid.UUID          `db:"id"`
	UserId              uuid.UUID          `db:"id_user"`
	Height              int                `db:"height"`
	Photo               string             `db:"photo"`
	Wants               entities.Want      `db:"wants"`
	Lifestyle           entities.Lifestyle `db:"lifestyle"`
	TargetWeight        float64            `db:"target_weight"`
	TargetWorkoutsWeeks int                `db:"target_workouts_weeks"`
	TargetCaloriesDaily int                `db:"target_calories_daily"`
	CurrentWeight       sql.NullFloat64    `db:"current_weight"`
}

func NewUserParamsRow(userParams *entities.UserParams) *UserParamsRow {
	return &UserParamsRow{
		ID:                  userParams.ID(),
		UserId:              userParams.UserID(),
		Height:              userParams.Height(),
		Photo:               userParams.Photo(),
		Wants:               userParams.Want(),
		TargetWeight:        userParams.TargetWeight(),
		TargetCaloriesDaily: userParams.TargetCaloriesDaily(),
		TargetWorkoutsWeeks: userParams.TargetWorkoutsWeeks(),
		Lifestyle:           userParams.Lifestyle(),
	}
}

func (u *UserParamsRow) ToEntity() *entities.UserParams {
	var weight float64
	if u.CurrentWeight.Valid {
		weight = u.CurrentWeight.Float64
	}
	return entities.NewUserParams(
		entities.WithUserParamsRestoreSpec(entities.UserParamsRestoreSpec{
			ID:                  u.ID,
			UserID:              u.UserId,
			Height:              u.Height,
			Photo:               u.Photo,
			Wants:               u.Wants,
			Lifestyle:           u.Lifestyle,
			TargetWeight:        u.TargetWeight,
			TargetWorkoutsWeeks: u.TargetWorkoutsWeeks,
			TargetCaloriesDaily: u.TargetCaloriesDaily,
			CurrentWeight:       weight,
		}),
	)
}
