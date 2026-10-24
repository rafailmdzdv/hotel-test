package domain

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input data")
	ErrPasswordLength = errors.New("password too short or long. use at least 6 chars and less that 72 chars")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrDateConflict  = errors.New("check-out date must be after check-in date")
)
