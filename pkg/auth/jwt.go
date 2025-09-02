package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

type TokenManager interface {
	GenerateTokenPair(userID int, role string) (string, string, error)
	ParseAccessToken(accessToken string) (*TokenClaims, error)
	ParseRefreshToken(refreshToken string) (*TokenClaims, error)
	GetRefreshTokenTTL() time.Duration
}

type JWTManager struct {
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (m *JWTManager) GenerateTokenPair(userID int, role string) (string, string, error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(m.accessTTL).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(m.accessSecret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(m.refreshTTL).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(m.refreshSecret))
	if err != nil {
		return "", "", err
	}

	return signedAccessToken, signedRefreshToken, nil
}

func (m *JWTManager) ParseAccessToken(accessToken string) (*TokenClaims, error) {
	return m.parseToken(accessToken, m.accessSecret)
}

func (m *JWTManager) ParseRefreshToken(refreshToken string) (*TokenClaims, error) {
	return m.parseToken(refreshToken, m.refreshSecret)
}

func (m *JWTManager) GetRefreshTokenTTL() time.Duration {
	return m.refreshTTL
}

func (m *JWTManager) parseToken(token, secret string) (*TokenClaims, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user_id claim")
	}
	userID := int(userIDFloat)

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid role claim")
	}

	return &TokenClaims{
		UserID: userID,
		Role:   role,
	}, nil
}
