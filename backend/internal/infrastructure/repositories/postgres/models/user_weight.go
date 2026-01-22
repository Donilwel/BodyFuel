package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type UserWeightRow struct {
	ID     uuid.UUID `db:"id"`
	UserId uuid.UUID `db:"id_user"`
	Weight float64   `db:"weight"`
	Date   time.Time `db:"date"`
}

func (u *UserWeightRow) ToEntity() *entities.UserWeight {
	return entities.NewUserWeight(entities.WithUserWeightRestoreSpec(entities.UserWeightRestoreSpec{
		ID:     u.ID,
		UserID: u.UserId,
		Weight: u.Weight,
		Date:   u.Date,
	}))
}
func NewUserWeightRow(userWeight *entities.UserWeight) *UserWeightRow {
	return &UserWeightRow{
		ID:     userWeight.ID(),
		UserId: userWeight.UserID(),
		Weight: userWeight.Weight(),
		Date:   userWeight.Date(),
	}
}
