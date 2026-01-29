package errors

import "errors"

var (
	ErrUnknownExerciseLevel   = errors.New("unknown exercise level")
	ErrUnknownExerciseType    = errors.New("unknown exercise type")
	ErrUnknownExercisePlace   = errors.New("unknown exercise place")
	ErrUnknownExerciseStatus  = errors.New("unknown exercise status")
	ErrExerciseNotFound       = errors.New("exercise not found")
	ErrInvalidExerciseData    = errors.New("invalid exercise data")
	ErrExerciseAlreadyDeleted = errors.New("exercise already deleted")
	ErrExerciseAlreadyExists  = errors.New("exercise already exists")
)
