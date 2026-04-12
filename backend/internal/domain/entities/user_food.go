package entities

import (
	"github.com/google/uuid"
	"time"
)

type MealType string

const (
	MealTypeBreakfast MealType = "breakfast"
	MealTypeLunch     MealType = "lunch"
	MealTypeDinner    MealType = "dinner"
	MealTypeSnack     MealType = "snack"
)

type UserFood struct {
	id          uuid.UUID
	userID      uuid.UUID
	description string
	calories    int
	protein     float64
	carbs       float64
	fat         float64
	mealType    MealType
	photoURL    string
	date        time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func (f *UserFood) ID() uuid.UUID        { return f.id }
func (f *UserFood) UserID() uuid.UUID    { return f.userID }
func (f *UserFood) Description() string  { return f.description }
func (f *UserFood) Calories() int        { return f.calories }
func (f *UserFood) Protein() float64     { return f.protein }
func (f *UserFood) Carbs() float64       { return f.carbs }
func (f *UserFood) Fat() float64         { return f.fat }
func (f *UserFood) MealType() MealType   { return f.mealType }
func (f *UserFood) PhotoURL() string     { return f.photoURL }
func (f *UserFood) Date() time.Time      { return f.date }
func (f *UserFood) CreatedAt() time.Time { return f.createdAt }
func (f *UserFood) UpdatedAt() time.Time { return f.updatedAt }

type UserFoodOption func(f *UserFood)

func NewUserFood(opt UserFoodOption) *UserFood {
	f := new(UserFood)
	opt(f)
	return f
}

type UserFoodInitSpec struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Description string
	Calories    int
	Protein     float64
	Carbs       float64
	Fat         float64
	MealType    MealType
	PhotoURL    string
	Date        time.Time
}

type UserFoodRestoreSpec struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Description string
	Calories    int
	Protein     float64
	Carbs       float64
	Fat         float64
	MealType    MealType
	PhotoURL    string
	Date        time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func WithUserFoodInitSpec(s UserFoodInitSpec) UserFoodOption {
	return func(f *UserFood) {
		f.id = s.ID
		f.userID = s.UserID
		f.description = s.Description
		f.calories = s.Calories
		f.protein = s.Protein
		f.carbs = s.Carbs
		f.fat = s.Fat
		f.mealType = s.MealType
		f.photoURL = s.PhotoURL
		f.date = s.Date
		f.createdAt = time.Now()
		f.updatedAt = time.Now()
	}
}

func WithUserFoodRestoreSpec(s UserFoodRestoreSpec) UserFoodOption {
	return func(f *UserFood) {
		f.id = s.ID
		f.userID = s.UserID
		f.description = s.Description
		f.calories = s.Calories
		f.protein = s.Protein
		f.carbs = s.Carbs
		f.fat = s.Fat
		f.mealType = s.MealType
		f.photoURL = s.PhotoURL
		f.date = s.Date
		f.createdAt = s.CreatedAt
		f.updatedAt = s.UpdatedAt
	}
}

type UserFoodUpdateParams struct {
	Description *string
	Calories    *int
	Protein     *float64
	Carbs       *float64
	Fat         *float64
	MealType    *MealType
	PhotoURL    *string
	Date        *time.Time
}

func (f *UserFood) Update(p UserFoodUpdateParams) {
	if p.Description != nil {
		f.description = *p.Description
	}
	if p.Calories != nil {
		f.calories = *p.Calories
	}
	if p.Protein != nil {
		f.protein = *p.Protein
	}
	if p.Carbs != nil {
		f.carbs = *p.Carbs
	}
	if p.Fat != nil {
		f.fat = *p.Fat
	}
	if p.MealType != nil {
		f.mealType = *p.MealType
	}
	if p.PhotoURL != nil {
		f.photoURL = *p.PhotoURL
	}
	if p.Date != nil {
		f.date = *p.Date
	}
	f.updatedAt = time.Now()
}
