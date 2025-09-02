package domain

import "time"

type OTPSession struct {
	ID         string    `json:"id"`
	Phone      string    `json:"phone"`
	Code       string    `json:"-"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsVerified bool      `json:"is_verified"`
	IsUsed     bool      `json:"is_used"`
	CreatedAt  time.Time `json:"created_at"`
}

type OTPRequest struct {
	Phone string `json:"phone" binding:"required,len=11,numeric"`
}

type OTPVerifyRequest struct {
	Phone string `json:"phone" binding:"required,len=11,numeric"`
	Code  string `json:"code" binding:"required,len=4,numeric"`
	ID    string `json:"id" binding:"required"`
}

type OTPRequestResponse struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Phone string `json:"phone"`
	From  string `json:"from"`
}

type OTPVerifyResponse struct {
	Phone          string `json:"phone"`
	Validated      bool   `json:"validated"`
	ValidationDate *int64 `json:"validation_date"`
}

type OTPStatusResponse struct {
	Phone          string `json:"phone"`
	Validated      bool   `json:"validated"`
	ValidationDate *int64 `json:"validation_date"`
}

type OTPErrorType string

const (
	OTPErrorNone          OTPErrorType = ""
	OTPErrorNotFound      OTPErrorType = "session_not_found"
	OTPErrorPhoneMismatch OTPErrorType = "phone_mismatch"
	OTPErrorExpired       OTPErrorType = "expired"
	OTPErrorAlreadyUsed   OTPErrorType = "already_used"
	OTPErrorInvalidCode   OTPErrorType = "invalid_code"
)

type OTPAuthResponse struct {
	RequiresRegistration bool         `json:"requires_registration"`
	AccessToken          *string      `json:"access_token,omitempty"`
	RefreshToken         *string      `json:"refresh_token,omitempty"`
	User                 *User        `json:"user,omitempty"`
	Message              string       `json:"message"`
	ErrorType            OTPErrorType `json:"error_type,omitempty"`
}

type OTPService interface {
	RequestOTP(phone string) (*OTPRequestResponse, error)

	VerifyOTP(id, code string) (*OTPVerifyResponse, error)

	CheckStatus(id string) (*OTPStatusResponse, error)
}

type OTPUseCase interface {
	RequestOTP(phone string) (*OTPRequestResponse, error)
	VerifyOTP(phone, code string) (bool, error)
	VerifyOTPAndAuthenticate(id, phone, code string) (*OTPAuthResponse, error)
	CheckStatus(id string) (*OTPStatusResponse, error)
}

type OTPRepository interface {
	CreateSession(session *OTPSession) error

	GetSessionByID(id string) (*OTPSession, error)

	GetSessionByPhone(phone string) (*OTPSession, error)

	UpdateSession(session *OTPSession) error

	DeleteSession(id string) error

	DeleteExpiredSessions() error
}
