package entities

import (
	"github.com/google/uuid"
	"time"
)

type UserWeight struct {
	id     uuid.UUID
	userId uuid.UUID
	weight float64
	date   time.Time
}

func (u *UserWeight) ID() uuid.UUID {
	return u.id
}

func (u *UserWeight) UserID() uuid.UUID {
	return u.userId
}

func (u *UserWeight) Weight() float64 {
	return u.weight
}

func (u *UserWeight) Date() time.Time {
	return u.date
}

type UserWeightOption func(u *UserWeight)

func NewUserWeight(opt UserWeightOption) *UserWeight {
	u := new(UserWeight)

	opt(u)

	return u
}

type UserWeightRestoreSpec struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Weight float64
	Date   time.Time
}

type UserWeightInitSpec struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Weight float64
	Date   time.Time
}

func WithUserWeightRestoreSpec(spec UserWeightRestoreSpec) UserWeightOption {
	return func(u *UserWeight) {
		u.id = spec.ID
		u.userId = spec.UserID
		u.weight = spec.Weight
		u.date = spec.Date
	}
}

func WithUserWeightInitSpec(s UserWeightInitSpec) UserWeightOption {
	return func(u *UserWeight) {
		u.id = s.ID
		u.userId = s.UserID
		u.weight = s.Weight
		u.date = s.Date
	}
}

type UserWeightUpdateParams struct {
	Weight *float64
	Date   *time.Time
}

func (ui *UserWeight) Update(p UserWeightUpdateParams) {
	if p.Weight != nil {
		ui.weight = *p.Weight
	}
}
