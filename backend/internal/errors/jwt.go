package errors

import "errors"

var (
	ErrHashedPassword  = errors.New("password hashed is incorrect")
	ErrTokenGeneration = errors.New("token generation failed")
)
