package errors

import "errors"

var (
	ErrUserInfoNotFound   = errors.New("user info not found")
	ErrUserAlreadyExists  = errors.New("user already exist")
	ErrInvalidCredentials = errors.New("password is incorrect")
)
