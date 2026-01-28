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

func (c Lifestyle) ToLevelPreparation() (LevelPreparation, error) {
	switch c {
	case NotActive:
		return Beginner, nil
	case Active:
		return Medium, nil
	case Sportive:
		return Sportsman, nil
	default:
		return "", fmt.Errorf("%s : %s", "mathing Lifestyle to LevelPreparation unknown type", c)
	}
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
	id                  uuid.UUID
	userId              uuid.UUID
	height              int
	photo               string
	wants               Want
	lifestyle           Lifestyle
	targetWeight        float64
	targetWorkoutsWeeks int
	targetCaloriesDaily int
	currentWeight       float64
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

func (u *UserParams) Photo() string {
	return u.photo
}

func (u *UserParams) Want() Want {
	return u.wants
}

func (u *UserParams) Lifestyle() Lifestyle {
	return u.lifestyle
}

func (u *UserParams) CurrentWeight() float64 {
	return u.currentWeight
}

func (u *UserParams) TargetWeight() float64 {
	return u.targetWeight
}

func (u *UserParams) TargetWorkoutsWeeks() int {
	return u.targetWorkoutsWeeks
}

func (u *UserParams) TargetCaloriesDaily() int {
	return u.targetCaloriesDaily
}

type UserParamsOption func(u *UserParams)

func NewUserParams(opt UserParamsOption) *UserParams {
	u := new(UserParams)

	opt(u)

	return u
}

type UserParamsRestoreSpec struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	Height              int
	Photo               string
	Wants               Want
	Lifestyle           Lifestyle
	TargetWeight        float64
	TargetWorkoutsWeeks int
	TargetCaloriesDaily int
	CurrentWeight       float64
}

type UserParamsInitSpec struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	Height              int
	Photo               string
	Wants               Want
	Lifestyle           Lifestyle
	TargetWeight        float64
	TargetWorkoutsWeeks int
	TargetCaloriesDaily int
}

func WithUserParamsRestoreSpec(spec UserParamsRestoreSpec) UserParamsOption {
	return func(u *UserParams) {
		u.id = spec.ID
		u.userId = spec.UserID
		u.height = spec.Height
		u.photo = spec.Photo
		u.wants = spec.Wants
		u.lifestyle = spec.Lifestyle
		u.targetWorkoutsWeeks = spec.TargetWorkoutsWeeks
		u.targetCaloriesDaily = spec.TargetCaloriesDaily
		u.targetWeight = spec.TargetWeight
		u.currentWeight = spec.CurrentWeight
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
		u.targetWorkoutsWeeks = spec.TargetWorkoutsWeeks
		u.targetCaloriesDaily = spec.TargetCaloriesDaily
		u.targetWeight = spec.TargetWeight
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
	if p.TargetCaloriesDaily != nil {
		r.targetCaloriesDaily = *p.TargetCaloriesDaily
	}
	if p.TargetWeight != nil {
		r.targetWeight = *p.TargetWeight
	}
	if p.TargetWorkoutsWeeks != nil {
		r.targetWorkoutsWeeks = *p.TargetWorkoutsWeeks
	}
}

type UserParamsUpdateParams struct {
	Height              *int
	Photo               *string
	Wants               *Want
	Lifestyle           *Lifestyle
	TargetWeight        *float64
	TargetWorkoutsWeeks *int
	TargetCaloriesDaily *int
}
