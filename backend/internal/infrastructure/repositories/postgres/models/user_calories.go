package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type UserCaloriesRow struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	Calories    int       `db:"calories"`
	Description string    `db:"description"`
	Date        time.Time `db:"date"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewUserCaloriesRow(u *entities.UserCalories) *UserCaloriesRow {
	return &UserCaloriesRow{
		ID:          u.ID(),
		UserID:      u.UserID(),
		Calories:    u.Calories(),
		Description: u.Description(),
		Date:        u.Date(),
		CreatedAt:   u.CreatedAt(),
		UpdatedAt:   u.UpdatedAt(),
	}
}

func (r *UserCaloriesRow) ToEntity() *entities.UserCalories {
	return entities.NewUserCalories(entities.WithUserCaloriesRestoreSpec(entities.UserCaloriesRestoreSpec{
		ID:          r.ID,
		UserID:      r.UserID,
		Calories:    r.Calories,
		Description: r.Description,
		Date:        r.Date,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}))
}
