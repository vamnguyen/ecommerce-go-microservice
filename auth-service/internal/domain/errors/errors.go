package errors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrWeakPassword       = errors.New("password is too weak")
	
	ErrAccountLocked   = errors.New("account is locked")
	ErrAccountInactive = errors.New("account is inactive")
	
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token expired")
	ErrTokenRevoked   = errors.New("token revoked")
	ErrMissingToken   = errors.New("missing token")
	
	ErrInvalidInput    = errors.New("invalid input")
	ErrValidationError = errors.New("validation error")
	
	ErrInternalServer = errors.New("internal server error")
	ErrDatabase       = errors.New("database error")
)
