package errors

import "errors"

var (
	ErrUserWeightNotFound       = errors.New("user weight not found")
	ErrUnknownUserWeightWant    = errors.New("unknown user weight type of field want")
	ErrUserWeightAlreadyExists  = errors.New("user weight with user_id already exist")
	ErrUnknownUserWeight        = errors.New("unknown user weight type of field lifestyle")
	ErrUserWeightAlreadyDeleted = errors.New("user weight with user_id already deleted")
)
