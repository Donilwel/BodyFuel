package entities

import (
	"backend/internal/errors"
	"fmt"
	"github.com/google/uuid"
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

func (c Want) ToString() string {
	return string(c)
}

func (c Lifestyle) ToString() string {
	return string(c)
}

func ToWant(s string) (Want, error) {
	switch s {
	case LoseWeight.ToString():
		return LoseWeight, nil
	case BuildMuscle.ToString():
		return BuildMuscle, nil
	case StayFit.ToString():
		return StayFit, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownUserParamsWant, s)
	}
}

func ToLifestyle(s string) (Lifestyle, error) {
	switch s {
	case NotActive.ToString():
		return NotActive, nil
	case Active.ToString():
		return Active, nil
	case Sportive.ToString():
		return Sportive, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownUserParams, s)
	}
}

type UserParams struct {
	id        uuid.UUID
	userId    uuid.UUID
	height    int
	photo     string
	wants     Want
	lifestyle Lifestyle
}

//type WantDay struct {
//	date time.Time
//	want Want
//}

func (u *UserParams) ID() uuid.UUID {
	return u.id
}

func (u *UserParams) UserID() uuid.UUID {
	return u.userId
}

func (u *UserParams) Height() int {
	return u.height
}

func (u *UserParams) Photo() string {
	return u.photo
}

func (u *UserParams) Want() Want {
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
	Photo     string
	Wants     Want
	Lifestyle Lifestyle
}

type UserParamsInitSpec struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Height    int
	Photo     string
	Wants     Want
	Lifestyle Lifestyle
}

func WithUserParamsRestoreSpec(spec UserParamsRestoreSpec) UserParamsOption {
	return func(u *UserParams) {
		u.id = spec.ID
		u.userId = spec.UserID
		u.height = spec.Height
		u.photo = spec.Photo
		u.wants = spec.Wants
		u.lifestyle = spec.Lifestyle
	}
}

func WithUserParamsInitSpec(spec UserParamsInitSpec) UserParamsOption {
	return func(u *UserParams) {
		u.id = spec.ID
		u.userId = spec.UserID
		u.height = spec.Height
		u.photo = spec.Photo
		u.wants = spec.Wants
		u.lifestyle = spec.Lifestyle
	}
}

func (r *UserParams) Update(p UserParamsUpdateParams) {
	if p.Height != nil {
		r.height = *p.Height
	}
	if p.Photo != nil {
		r.photo = *p.Photo
	}
	if p.Wants != nil {
		r.wants = *p.Wants
	}
	if p.Lifestyle != nil {
		r.lifestyle = *p.Lifestyle
	}
}

type UserParamsUpdateParams struct {
	Height    *int
	Photo     *string
	Wants     *Want
	Lifestyle *Lifestyle
}
