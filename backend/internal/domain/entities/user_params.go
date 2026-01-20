package entities

import (
	"github.com/google/uuid"
	"time"
)

type Lifestyle string

func (l Lifestyle) String() string {
	return string(l)
}

type Want string

func (w Want) String() string {
	return string(w)
}

const (
	LoseWeight  Want = "lose_weight"
	BuildMuscle Want = "build_muscle"
	StayFit     Want = "stay_fit"

	NotActive Lifestyle = "not_active"
	Active    Lifestyle = "active"
	Sportive  Lifestyle = "sportive"
)

type UserParams struct {
	id        uuid.UUID
	userId    uuid.UUID
	height    int
	weight    []WeightDay
	photo     string
	wants     []WantDay
	lifestyle Lifestyle
}

type WeightDay struct {
	date   time.Time
	weight float64
}

type WantDay struct {
	date time.Time
	want Want
}

func (u *UserParams) ID() uuid.UUID {
	return u.id
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

func (u *UserParams) Lifestyle() Lifestyle {
	return u.lifestyle
}

type UserParamsOption func(u *UserParams)

func NewUserParams(opt UserParamsOption) *UserParams {
	u := new(UserParams)

	opt(u)

	return u
}

type UserParamsRestoreSpec struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Height    int
	Weight    []WeightDay
	Photo     string
	Wants     []WantDay
	Lifestyle Lifestyle
}

type UserParamsInitSpec struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Height    int
	Weight    []WeightDay
	Photo     string
	Wants     []WantDay
	Lifestyle Lifestyle
}

func WithUserParamsRestoreSpec(spec UserParamsRestoreSpec) UserParamsOption {
	return func(u *UserParams) {
		u.id = spec.ID
		u.userId = spec.UserID
		u.height = spec.Height
		u.weight = spec.Weight
		u.photo = spec.Photo
		u.wants = spec.Wants
		u.lifestyle = spec.Lifestyle
	}
}
