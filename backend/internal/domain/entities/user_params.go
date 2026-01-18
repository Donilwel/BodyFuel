package entities

import (
	"github.com/google/uuid"
	"time"
)

type Want string

func (w Want) String() string {
	return string(w)
}

const (
	LoseWeight Want = "lose_weight"
	BuildMuscle Want = "build_muscle"
	StayFit     Want = "stay_fit"
)


type UserParams struct {
	userId uuid.UUID
	height int
	weight []WeightDay
	photo  string
	wants  []WantDay
}

type WeightDay struct {
	date   time.Time
	weight float64
}

type WantDay struct {
	date time.Time
	want Want
}

func (u *UserParams) UserID() uuid.UUID {
	return u.userId
}

func (u *UserParams) Height() int {
	return u.height
}

func (u *UserParams) Weight() []WeightDay {
	return u.weight
}

func (u *UserParams) Photo() string {
	return u.photo
}

func (u *UserParams) Wants() []WantDay {
	return u.wants
}

type UserParamsOption func(u *UserParams)

func NewUserParams(opt UserParamsOption) *UserParams {
	u := new(UserParams)

	opt(u)

	return u
}

type UserParamsRestoreSpec struct {
	UserID uuid.UUID
	Height int
	Weight []WeightDay
	Photo  string
	Wants  []WantDay
}



func WithUserParamsRestoreSpec func(spec UserParamsRestoreSpec) UserParamsOption {
	return func(u *UserParams) {
		u.userId = spec.UserID
		u.height = spec.Height
		u.weight = spec.Weight
		u.photo = spec.Photo
		u.wants = spec.Wants
	}
}


