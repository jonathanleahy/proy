package user

import "errors"

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidUserID is returned when the user ID is invalid
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrInvalidUserData is returned when user data is invalid
	ErrInvalidUserData = errors.New("invalid user data")

	// ErrExternalServiceUnavailable is returned when the external service is unavailable
	ErrExternalServiceUnavailable = errors.New("external service unavailable")
)
