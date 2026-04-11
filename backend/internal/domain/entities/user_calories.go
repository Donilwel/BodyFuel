package entities

import (
	"github.com/google/uuid"
	"time"
)

type UserCalories struct {
	id          uuid.UUID
	userID      uuid.UUID
	calories    int
	description string
	date        time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func (u *UserCalories) ID() uuid.UUID        { return u.id }
func (u *UserCalories) UserID() uuid.UUID     { return u.userID }
func (u *UserCalories) Calories() int         { return u.calories }
func (u *UserCalories) Description() string   { return u.description }
func (u *UserCalories) Date() time.Time       { return u.date }
func (u *UserCalories) CreatedAt() time.Time  { return u.createdAt }
func (u *UserCalories) UpdatedAt() time.Time  { return u.updatedAt }

type UserCaloriesOption func(u *UserCalories)

func NewUserCalories(opt UserCaloriesOption) *UserCalories {
	u := new(UserCalories)
	opt(u)
	return u
}

type UserCaloriesInitSpec struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Calories    int
	Description string
	Date        time.Time
}

type UserCaloriesRestoreSpec struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Calories    int
	Description string
	Date        time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func WithUserCaloriesInitSpec(s UserCaloriesInitSpec) UserCaloriesOption {
	return func(u *UserCalories) {
		u.id = s.ID
		u.userID = s.UserID
		u.calories = s.Calories
		u.description = s.Description
		u.date = s.Date
		u.createdAt = time.Now()
		u.updatedAt = time.Now()
	}
}

func WithUserCaloriesRestoreSpec(s UserCaloriesRestoreSpec) UserCaloriesOption {
	return func(u *UserCalories) {
		u.id = s.ID
		u.userID = s.UserID
		u.calories = s.Calories
		u.description = s.Description
		u.date = s.Date
		u.createdAt = s.CreatedAt
		u.updatedAt = s.UpdatedAt
	}
}

type UserCaloriesUpdateParams struct {
	Calories    *int
	Description *string
	Date        *time.Time
}

func (u *UserCalories) Update(p UserCaloriesUpdateParams) {
	if p.Calories != nil {
		u.calories = *p.Calories
	}
	if p.Description != nil {
		u.description = *p.Description
	}
	if p.Date != nil {
		u.date = *p.Date
	}
	u.updatedAt = time.Now()
}
