package errors

import "errors"

var (
	ErrWorkoutsExerciseNotFound       = errors.New("workouts exercise not found")
	ErrWorkoutsExerciseAlreadyDeleted = errors.New("workouts exercise already deleted")
)
