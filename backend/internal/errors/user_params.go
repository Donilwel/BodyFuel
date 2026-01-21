package errors

import "errors"

var (
	ErrUserParamsNotFound    = errors.New("user params not found")
	ErrUnknownUserParamsWant = errors.New("unknown user params type of field want")
	ErrUnknownUserParams     = errors.New("unknown user params type of field lifestyle")
)
