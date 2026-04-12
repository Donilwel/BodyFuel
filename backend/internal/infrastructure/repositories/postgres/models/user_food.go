package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type UserFoodRow struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	Description string    `db:"description"`
	Calories    int       `db:"calories"`
	Protein     float64   `db:"protein"`
	Carbs       float64   `db:"carbs"`
	Fat         float64   `db:"fat"`
	MealType    string    `db:"meal_type"`
	PhotoURL    string    `db:"photo_url"`
	Date        time.Time `db:"date"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewUserFoodRow(f *entities.UserFood) *UserFoodRow {
	return &UserFoodRow{
		ID:          f.ID(),
		UserID:      f.UserID(),
		Description: f.Description(),
		Calories:    f.Calories(),
		Protein:     f.Protein(),
		Carbs:       f.Carbs(),
		Fat:         f.Fat(),
		MealType:    string(f.MealType()),
		PhotoURL:    f.PhotoURL(),
		Date:        f.Date(),
		CreatedAt:   f.CreatedAt(),
		UpdatedAt:   f.UpdatedAt(),
	}
}

func (r *UserFoodRow) ToEntity() *entities.UserFood {
	return entities.NewUserFood(entities.WithUserFoodRestoreSpec(entities.UserFoodRestoreSpec{
		ID:          r.ID,
		UserID:      r.UserID,
		Description: r.Description,
		Calories:    r.Calories,
		Protein:     r.Protein,
		Carbs:       r.Carbs,
		Fat:         r.Fat,
		MealType:    entities.MealType(r.MealType),
		PhotoURL:    r.PhotoURL,
		Date:        r.Date,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}))
}
