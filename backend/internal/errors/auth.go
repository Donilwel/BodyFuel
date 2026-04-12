package errors

import "errors"

var (
	ErrTokenExpired                = errors.New("token is expired")
	ErrInvalidVerificationCode     = errors.New("verification code is invalid")
	ErrVerificationCodeExpired     = errors.New("verification code is expired")
	ErrVerificationCodeAlreadyUsed = errors.New("verification code already used")
)
