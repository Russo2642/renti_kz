package errors

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	ErrInvalidCredentials = "invalid_credentials"
	ErrInvalidToken       = "invalid_token"
	ErrExpiredToken       = "expired_token"
	ErrUserNotFound       = "user_not_found"
	ErrAccountNotActive   = "account_not_active"

	ErrUserAlreadyExists  = "user_already_exists"
	ErrEmailAlreadyExists = "email_already_exists"
	ErrPhoneAlreadyExists = "phone_already_exists"
	ErrInvalidPassword    = "invalid_password"

	ErrDatabaseError = "database_error"
	ErrStorageError  = "storage_error"
	ErrNotFound      = "not_found"

	ErrValidation = "validation_error"
	ErrBadRequest = "bad_request"

	ErrInternal     = "internal_error"
	ErrForbidden    = "forbidden"
	ErrUnauthorized = "unauthorized"
)

type AppError struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(errType string, message string, statusCode int, err error) *AppError {
	return &AppError{
		Type:       errType,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

func NewInvalidCredentials(message string) *AppError {
	return New(ErrInvalidCredentials, message, http.StatusUnauthorized, nil)
}

func NewInvalidToken(message string) *AppError {
	return New(ErrInvalidToken, message, http.StatusUnauthorized, nil)
}

func NewExpiredToken(message string) *AppError {
	return New(ErrExpiredToken, message, http.StatusUnauthorized, nil)
}

func NewUserNotFound(message string) *AppError {
	return New(ErrUserNotFound, message, http.StatusNotFound, nil)
}

func NewUserAlreadyExists(message string) *AppError {
	return New(ErrUserAlreadyExists, message, http.StatusConflict, nil)
}

func NewEmailAlreadyExists(message string) *AppError {
	return New(ErrEmailAlreadyExists, message, http.StatusConflict, nil)
}

func NewPhoneAlreadyExists(message string) *AppError {
	return New(ErrPhoneAlreadyExists, message, http.StatusConflict, nil)
}

func NewInvalidPassword(message string) *AppError {
	return New(ErrInvalidPassword, message, http.StatusBadRequest, nil)
}

func NewDatabaseError(message string, err error) *AppError {
	return New(ErrDatabaseError, message, http.StatusInternalServerError, err)
}

func NewStorageError(message string, err error) *AppError {
	return New(ErrStorageError, message, http.StatusInternalServerError, err)
}

func NewNotFound(message string) *AppError {
	return New(ErrNotFound, message, http.StatusNotFound, nil)
}

func NewValidationError(message string) *AppError {
	return New(ErrValidation, message, http.StatusBadRequest, nil)
}

func NewBadRequestError(message string) *AppError {
	return New(ErrBadRequest, message, http.StatusBadRequest, nil)
}

func NewInternalError(message string, err error) *AppError {
	return New(ErrInternal, message, http.StatusInternalServerError, err)
}

func NewForbiddenError(message string) *AppError {
	return New(ErrForbidden, message, http.StatusForbidden, nil)
}

func NewUnauthorizedError(message string) *AppError {
	return New(ErrUnauthorized, message, http.StatusUnauthorized, nil)
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func Is(err error, errType string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

func Wrap(err error, errType string, message string, statusCode int) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if message != "" {
			appErr.Message = message
		}
		if errType != "" {
			appErr.Type = errType
		}
		if statusCode != 0 {
			appErr.StatusCode = statusCode
		}
		return appErr
	}

	if message == "" {
		message = err.Error()
	}
	return New(errType, message, statusCode, err)
}

func FormatError(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Err != nil {
			return fmt.Sprintf("[%s] %s: %v", appErr.Type, appErr.Message, appErr.Err)
		}
		return fmt.Sprintf("[%s] %s", appErr.Type, appErr.Message)
	}
	return err.Error()
}
