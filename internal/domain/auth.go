package domain

import "time"

type Credentials struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Phone        string       `json:"phone" binding:"required,len=11,numeric"`
	FirstName    string       `json:"first_name" binding:"required"`
	LastName     string       `json:"last_name" binding:"required"`
	Email        string       `json:"email" binding:"required,email"`
	CityID       int          `json:"city_id" binding:"required"`
	IIN          string       `json:"iin" binding:"required,len=12,numeric"`
	DocumentType DocumentType `json:"-"`
}

type PreRegistrationRequest struct {
	Phone     string `json:"phone" binding:"required,len=11,numeric"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	CityID    int    `json:"city_id" binding:"required"`
	IIN       string `json:"iin" binding:"required,len=12,numeric"`
}

type CompleteRegistrationRequest struct {
	Phone     string `json:"phone" binding:"required,len=11,numeric"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	CityID    int    `json:"city_id" binding:"required"`
	IIN       string `json:"iin" binding:"required,len=12,numeric"`
	OTPCode   string `json:"otp_code" binding:"required"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required,len=11,numeric"`
	Password string `json:"password" binding:"required"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Session struct {
	RefreshToken string    `json:"refresh_token"`
	UserID       int       `json:"user_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuthUseCase interface {
	SignIn(phone, password string) (*Tokens, *User, error)
	RefreshTokens(refreshToken string) (*Tokens, error)
	SignOut(refreshToken string) error
	SignOutAll(userID int) error
	GetUserFromToken(accessToken string) (*User, error)
}

type AuthRepository interface {
	CreateSession(session Session) error
	GetSessionByToken(refreshToken string) (*Session, error)
	DeleteSession(refreshToken string) error
	DeleteAllUserSessions(userID int) error
}
