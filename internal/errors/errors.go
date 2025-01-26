package errors

import "errors"

var (
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrInvalidToken              = errors.New("invalid token")
	ErrRegistrationDisabled      = errors.New("registration is disabled")
	ErrNotFound                  = errors.New("not found")
	ErrNotAllowed                = errors.New("not allowed")
	ErrSomethingWentWrong        = errors.New("something went wrong")
	ErrUsernameAlreadyRegistered = errors.New("username is already registered")
)
