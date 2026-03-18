package entities

import (
	"backend/internal/errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ExerciseStatus string

func (s ExerciseStatus) String() string {
	return string(s)
}

const (
	ExerciseStatusPending    ExerciseStatus = "pending"
	ExerciseStatusInProgress ExerciseStatus = "in_progress"
	ExerciseStatusCompleted  ExerciseStatus = "completed"
	ExerciseStatusSkipped    ExerciseStatus = "skipped"
)

func ToExerciseStatus(s string) (ExerciseStatus, error) {
	switch s {
	case ExerciseStatusPending.String():
		return ExerciseStatusPending, nil
	case ExerciseStatusInProgress.String():
		return ExerciseStatusInProgress, nil
	case ExerciseStatusCompleted.String():
		return ExerciseStatusCompleted, nil
	case ExerciseStatusSkipped.String():
		return ExerciseStatusSkipped, nil
	default:
		return "", fmt.Errorf("%w : %s", errors.ErrUnknownExerciseStatus, s)
	}
}

type WorkoutsExercise struct {
	workoutID       uuid.UUID
	exerciseID      uuid.UUID
	modifyReps      int
	modifyRelaxTime int
	calories        int
	status          ExerciseStatus
	orderIndex      int
	createdAt       time.Time
	updatedAt       time.Time
}

func (we *WorkoutsExercise) WorkoutID() uuid.UUID {
	return we.workoutID
}

func (we *WorkoutsExercise) ExerciseID() uuid.UUID {
	return we.exerciseID
}

func (we *WorkoutsExercise) ModifyReps() int {
	return we.modifyReps
}

func (we *WorkoutsExercise) ModifyRelaxTime() int {
	return we.modifyRelaxTime
}

func (we *WorkoutsExercise) Calories() int {
	return we.calories
}

func (we *WorkoutsExercise) Status() ExerciseStatus {
	return we.status
}

func (we *WorkoutsExercise) OrderIndex() int {
	return we.orderIndex
}

func (we *WorkoutsExercise) CreatedAt() time.Time {
	return we.createdAt
}

func (we *WorkoutsExercise) UpdatedAt() time.Time {
	return we.updatedAt
}

type WorkoutsExerciseOption func(we *WorkoutsExercise)

func NewWorkoutsExercise(opt WorkoutsExerciseOption) *WorkoutsExercise {
	we := new(WorkoutsExercise)
	opt(we)
	return we
}

type WorkoutsExerciseInitSpec struct {
	WorkoutID       uuid.UUID
	ExerciseID      uuid.UUID
	ModifyReps      int
	ModifyRelaxTime int
	Calories        int
	Status          ExerciseStatus
	OrderIndex      int
	UpdatedAt       time.Time
	CreatedAt       time.Time
}

type WorkoutsExerciseRestoreSpec struct {
	WorkoutID       uuid.UUID
	ExerciseID      uuid.UUID
	ModifyReps      int
	ModifyRelaxTime int
	Calories        int
	Status          ExerciseStatus
	OrderIndex      int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func WithWorkoutsExerciseInitSpec(s WorkoutsExerciseInitSpec) WorkoutsExerciseOption {
	return func(we *WorkoutsExercise) {
		we.workoutID = s.WorkoutID
		we.exerciseID = s.ExerciseID
		we.modifyReps = s.ModifyReps
		we.modifyRelaxTime = s.ModifyRelaxTime
		we.calories = s.Calories
		we.status = s.Status
		we.orderIndex = s.OrderIndex
		we.createdAt = s.CreatedAt
		we.updatedAt = s.CreatedAt
	}
}

func WithWorkoutsExerciseRestoreSpec(s WorkoutsExerciseRestoreSpec) WorkoutsExerciseOption {
	return func(we *WorkoutsExercise) {
		we.workoutID = s.WorkoutID
		we.exerciseID = s.ExerciseID
		we.modifyReps = s.ModifyReps
		we.modifyRelaxTime = s.ModifyRelaxTime
		we.calories = s.Calories
		we.status = s.Status
		we.orderIndex = s.OrderIndex
		we.createdAt = s.CreatedAt
		we.updatedAt = s.UpdatedAt
	}
}

type WorkoutsExerciseUpdateParams struct {
	ModifyReps      *int
	ModifyRelaxTime *int
	Calories        *int
	Status          *ExerciseStatus
	OrderIndex      *int
	UpdatedAt       *time.Time
}

func (we *WorkoutsExercise) Update(p WorkoutsExerciseUpdateParams) {
	if p.ModifyReps != nil {
		we.modifyReps = *p.ModifyReps
	}
	if p.ModifyRelaxTime != nil {
		we.modifyRelaxTime = *p.ModifyRelaxTime
	}
	if p.Calories != nil {
		we.calories = *p.Calories
	}
	if p.Status != nil {
		we.status = *p.Status
	}
	if p.OrderIndex != nil {
		we.orderIndex = *p.OrderIndex
	}
	if p.UpdatedAt != nil {
		we.updatedAt = *p.UpdatedAt
	}
}
