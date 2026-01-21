package errors

import "errors"

var (
	ErrUserParamsNotFound      = errors.New("user params not found")
	ErrUnknownUserParamsWant   = errors.New("unknown user params type of field want")
	ErrUserParamsAlreadyExists = errors.New("user params with user_id already exist")
	ErrUnknownUserParams       = errors.New("unknown user params type of field lifestyle")
)
