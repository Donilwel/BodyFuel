package entities

import (
	"backend/internal/errors"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type WorkoutsLevel string

func (l WorkoutsLevel) String() string {
	return string(l)
}

type WorkoutsStatus string

func (s WorkoutsStatus) String() string {
	return string(s)
}

const (
	WorkoutLight  WorkoutsLevel = "workout_light"
	WorkoutMiddle WorkoutsLevel = "workout_middle"
	WorkoutHard   WorkoutsLevel = "workout_hard"

	WorkoutStatusCreated  WorkoutsStatus = "workout_created"
	WorkoutStatusDone     WorkoutsStatus = "workout_done"
	WorkoutStatusInActive WorkoutsStatus = "workout_in_active"
	WorkoutStatusFailed   WorkoutsStatus = "workout_failed"
)

func (l WorkoutsLevel) ToString() string {
	return string(l)
}

func (s WorkoutsStatus) ToString() string {
	return string(s)
}

func ToWorkoutsLevel(s string) (WorkoutsLevel, error) {
	switch s {
	case WorkoutLight.ToString():
		return WorkoutLight, nil
	case WorkoutMiddle.ToString():
		return WorkoutMiddle, nil
	case WorkoutHard.ToString():
		return WorkoutHard, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownWorkoutsLevel, s)
	}
}

type Workout struct {
	id                 uuid.UUID
	userID             uuid.UUID
	level              WorkoutsLevel
	status             WorkoutsStatus
	totalCalories      int
	predictionCalories int
	duration           time.Duration
	createdAt          time.Time
	updatedAt          time.Time
}

func (w *Workout) ID() uuid.UUID {
	return w.id
}

func (w *Workout) UserID() uuid.UUID {
	return w.userID
}

func (w *Workout) Level() WorkoutsLevel {
	return w.level
}

func (w *Workout) Status() WorkoutsStatus {
	return w.status
}

func (w *Workout) TotalCalories() int {
	return w.totalCalories
}

func (w *Workout) PredictionCalories() int {
	return w.predictionCalories
}

func (w *Workout) Duration() time.Duration {
	return w.duration
}

func (w *Workout) CreatedAt() time.Time {
	return w.createdAt
}

func (w *Workout) UpdatedAt() time.Time {
	return w.updatedAt
}

type WorkoutOption func(w *Workout)

func NewWorkout(opt WorkoutOption) *Workout {
	w := new(Workout)
	opt(w)
	return w
}

type WorkoutInitSpec struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Level              WorkoutsLevel
	Status             WorkoutsStatus
	PredictionCalories int
	Duration           time.Duration
	CreatedAt          time.Time
}

type WorkoutRestoreSpec struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Level              WorkoutsLevel
	Status             WorkoutsStatus
	TotalCalories      int
	PredictionCalories int
	Duration           time.Duration
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func WithWorkoutInitSpec(s WorkoutInitSpec) WorkoutOption {
	return func(w *Workout) {
		w.id = s.ID
		w.userID = s.UserID
		w.level = s.Level
		w.status = s.Status
		w.predictionCalories = s.PredictionCalories
		w.duration = s.Duration
		w.createdAt = s.CreatedAt
		w.updatedAt = s.CreatedAt
	}
}

func WithWorkoutRestoreSpec(s WorkoutRestoreSpec) WorkoutOption {
	return func(w *Workout) {
		w.id = s.ID
		w.userID = s.UserID
		w.level = s.Level
		w.status = s.Status
		w.totalCalories = s.TotalCalories
		w.predictionCalories = s.PredictionCalories
		w.duration = s.Duration
		w.createdAt = s.CreatedAt
		w.updatedAt = s.UpdatedAt
	}
}

type WorkoutUpdateParams struct {
	Level              *WorkoutsLevel
	Status             *WorkoutsStatus
	TotalCalories      *int
	PredictionCalories *int
	Duration           *time.Duration
	UpdatedAt          *time.Time
}

func (w *Workout) Update(p WorkoutUpdateParams) {
	if p.Level != nil {
		w.level = *p.Level
	}
	if p.Status != nil {
		w.status = *p.Status
	}
	if p.TotalCalories != nil {
		w.totalCalories = *p.TotalCalories
	}
	if p.PredictionCalories != nil {
		w.predictionCalories = *p.PredictionCalories
	}
	if p.Duration != nil {
		w.duration = *p.Duration
	}
	if p.UpdatedAt != nil {
		w.updatedAt = *p.UpdatedAt
	}
}
