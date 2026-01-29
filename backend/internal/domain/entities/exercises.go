package entities

import (
	"backend/internal/errors"
	"fmt"
	"github.com/google/uuid"
)

type LevelPreparation string

func (l LevelPreparation) String() string {
	return string(l)
}

type ExerciseType string

func (e ExerciseType) String() string {
	return string(e)
}

type PlaceExercise string

func (p PlaceExercise) String() string {
	return string(p)
}

type ExerciseStatus string

func (e ExerciseStatus) String() string {
	return string(e)
}

const (
	Beginner  LevelPreparation = "beginner"
	Medium    LevelPreparation = "medium"
	Sportsman LevelPreparation = "sportsman"

	Cardio      ExerciseType = "cardio"
	UpperBody   ExerciseType = "upper_body"
	LowerBody   ExerciseType = "lower_body"
	FullBody    ExerciseType = "full_body"
	Flexibility ExerciseType = "flexibility"

	Street PlaceExercise = "street"
	Gym    PlaceExercise = "gym"
	Home   PlaceExercise = "home"

	NotStarted ExerciseStatus = "not_started"
	InProgress ExerciseStatus = "in_progress"
	Completed  ExerciseStatus = "completed"
	Skipped    ExerciseStatus = "skipped"
)

func (l LevelPreparation) ToString() string {
	return string(l)
}

func (e ExerciseType) ToString() string {
	return string(e)
}

func (p PlaceExercise) ToString() string {
	return string(p)
}

func (e ExerciseStatus) ToString() string {
	return string(e)
}

func ToLevelPreparation(s string) (LevelPreparation, error) {
	switch s {
	case Beginner.ToString():
		return Beginner, nil
	case Medium.ToString():
		return Medium, nil
	case Sportsman.ToString():
		return Sportsman, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownExerciseLevel, s)
	}
}

func ToExerciseType(s string) (ExerciseType, error) {
	switch s {
	case Cardio.ToString():
		return Cardio, nil
	case UpperBody.ToString():
		return UpperBody, nil
	case LowerBody.ToString():
		return LowerBody, nil
	case FullBody.ToString():
		return FullBody, nil
	case Flexibility.ToString():
		return Flexibility, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownExerciseType, s)
	}
}

func ToExerciseStatus(s string) (ExerciseStatus, error) {
	switch s {
	case NotStarted.ToString():
		return NotStarted, nil
	case InProgress.ToString():
		return InProgress, nil
	case Completed.ToString():
		return Completed, nil
	case Skipped.ToString():
		return Skipped, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownExerciseStatus, s)
	}
}

func ToPlaceExercise(s string) (PlaceExercise, error) {
	switch s {
	case Street.ToString():
		return Street, nil
	case Gym.ToString():
		return Gym, nil
	case Home.ToString():
		return Home, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownExercisePlace, s)
	}
}

type Exercise struct {
	id               uuid.UUID
	levelPreparation LevelPreparation
	name             string
	typeExercise     ExerciseType
	description      string
	baseCountReps    int
	steps            int
	linkGif          string
	placeExercise    PlaceExercise
	avgCaloriesPer   float64
	baseRelaxTime    int
}

func (e *Exercise) ID() uuid.UUID {
	return e.id
}

func (e *Exercise) LevelPreparation() LevelPreparation {
	return e.levelPreparation
}

func (e *Exercise) Name() string {
	return e.name
}

func (e *Exercise) TypeExercise() ExerciseType {
	return e.typeExercise
}

func (e *Exercise) Description() string {
	return e.description
}

func (e *Exercise) BaseCountReps() int {
	return e.baseCountReps
}

func (e *Exercise) Steps() int {
	return e.steps
}

func (e *Exercise) LinkGif() string {
	return e.linkGif
}

func (e *Exercise) PlaceExercise() PlaceExercise {
	return e.placeExercise
}

func (e *Exercise) AvgCaloriesPer() float64 {
	return e.avgCaloriesPer
}

func (e *Exercise) BaseRelaxTime() int {
	return e.baseRelaxTime
}

type ExerciseOption func(e *Exercise)

func NewExercise(opt ExerciseOption) *Exercise {
	e := new(Exercise)
	opt(e)
	return e
}

type ExerciseRestoreSpec struct {
	ID               uuid.UUID
	LevelPreparation LevelPreparation
	Name             string
	TypeExercise     ExerciseType
	Description      string
	BaseCountReps    int
	Steps            int
	LinkGif          string
	PlaceExercise    PlaceExercise
	AvgCaloriesPer   float64
	BaseRelaxTime    int
}

type ExerciseInitSpec struct {
	ID               uuid.UUID
	LevelPreparation LevelPreparation
	Name             string
	TypeExercise     ExerciseType
	Description      string
	BaseCountReps    int
	Steps            int
	LinkGif          string
	PlaceExercise    PlaceExercise
	AvgCaloriesPer   float64
	BaseRelaxTime    int
}

func WithExerciseRestoreSpec(spec ExerciseRestoreSpec) ExerciseOption {
	return func(e *Exercise) {
		e.id = spec.ID
		e.levelPreparation = spec.LevelPreparation
		e.name = spec.Name
		e.typeExercise = spec.TypeExercise
		e.description = spec.Description
		e.baseCountReps = spec.BaseCountReps
		e.steps = spec.Steps
		e.linkGif = spec.LinkGif
		e.placeExercise = spec.PlaceExercise
		e.avgCaloriesPer = spec.AvgCaloriesPer
		e.baseRelaxTime = spec.BaseRelaxTime
	}
}

func WithExerciseInitSpec(spec ExerciseInitSpec) ExerciseOption {
	return func(e *Exercise) {
		e.id = spec.ID
		e.levelPreparation = spec.LevelPreparation
		e.name = spec.Name
		e.typeExercise = spec.TypeExercise
		e.description = spec.Description
		e.baseCountReps = spec.BaseCountReps
		e.steps = spec.Steps
		e.linkGif = spec.LinkGif
		e.placeExercise = spec.PlaceExercise
		e.avgCaloriesPer = spec.AvgCaloriesPer
		e.baseRelaxTime = spec.BaseRelaxTime
	}
}

func (e *Exercise) Update(p ExerciseUpdateParams) {
	if p.LevelPreparation != nil {
		e.levelPreparation = *p.LevelPreparation
	}
	if p.Name != nil {
		e.name = *p.Name
	}
	if p.Description != nil {
		e.description = *p.Description
	}
	if p.BaseCountReps != nil {
		e.baseCountReps = *p.BaseCountReps
	}
	if p.Steps != nil {
		e.steps = *p.Steps
	}
	if p.LinkGif != nil {
		e.linkGif = *p.LinkGif
	}
	if p.AvgCaloriesPer != nil {
		e.avgCaloriesPer = *p.AvgCaloriesPer
	}
	if p.BaseRelaxTime != nil {
		e.baseRelaxTime = *p.BaseRelaxTime
	}
}

type ExerciseUpdateParams struct {
	LevelPreparation *LevelPreparation
	Name             *string
	Description      *string
	BaseCountReps    *int
	Steps            *int
	LinkGif          *string
	AvgCaloriesPer   *float64
	BaseRelaxTime    *int
}

func (e *Exercise) IsCardio() bool {
	return e.typeExercise == Cardio
}

func (e *Exercise) IsStrength() bool {
	return e.typeExercise == UpperBody ||
		e.typeExercise == LowerBody ||
		e.typeExercise == FullBody
}

func (e *Exercise) CalculateCalories(durationMinutes int, repetitions int) float64 {
	if e.IsCardio() {
		return e.avgCaloriesPer * float64(durationMinutes)
	} else if e.IsStrength() {
		return e.avgCaloriesPer * float64(repetitions)
	}
	return e.avgCaloriesPer * float64(durationMinutes)
}

func (e *Exercise) GetRecommendedRestTime() int {
	if e.baseRelaxTime > 0 {
		return e.baseRelaxTime
	}

	switch e.typeExercise {
	case Cardio:
		return 30
	case UpperBody, LowerBody, FullBody:
		return 60
	case Flexibility:
		return 15
	default:
		return 30
	}
}

func (e *Exercise) CanBeDoneAt(place PlaceExercise) bool {
	if e.placeExercise == "any" {
		return true
	}
	return e.placeExercise == place
}
