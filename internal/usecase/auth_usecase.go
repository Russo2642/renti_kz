package usecase

import (
	"errors"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
	"github.com/russo2642/renti_kz/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo         domain.UserRepository
	tokenManager     auth.TokenManager
	passwordSalt     string
	userCacheService *services.UserCacheService
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenManager auth.TokenManager,
	passwordSalt string,
	userCacheService *services.UserCacheService,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:         userRepo,
		tokenManager:     tokenManager,
		passwordSalt:     passwordSalt,
		userCacheService: userCacheService,
	}
}

func (uc *AuthUseCase) SignIn(phone, password string) (*domain.Tokens, *domain.User, error) {
	user, err := uc.userRepo.GetByPhone(phone)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user by phone: %w", err)
	}
	if user == nil {
		return nil, nil, errors.New("user not found")
	}

	// _, err = uc.hashPassword(password)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	// }

	if err := uc.comparePasswords(user.PasswordHash, password); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	accessToken, refreshToken, err := uc.tokenManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	tokens := &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return tokens, user, nil
}

func (uc *AuthUseCase) RefreshTokens(refreshToken string) (*domain.Tokens, error) {
	claims, err := uc.tokenManager.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userID := claims.UserID

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if uc.userCacheService != nil {
		_ = uc.userCacheService.InvalidateUser(userID)
	}

	accessToken, newRefreshToken, err := uc.tokenManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	tokens := &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}

	return tokens, nil
}

func (uc *AuthUseCase) SignOut(accessToken string) error {
	if uc.userCacheService != nil {
		return uc.userCacheService.InvalidateToken(accessToken)
	}
	return nil
}

func (uc *AuthUseCase) SignOutAll(userID int) error {
	if uc.userCacheService != nil {
		// Инвалидируем кэш пользователя
		if err := uc.userCacheService.InvalidateUser(userID); err != nil {
			return fmt.Errorf("ошибка инвалидации кэша пользователя: %w", err)
		}
		// Можно также очистить все токены, но это радикальная мера
		// return uc.userCacheService.InvalidateAllTokens()
	}
	return nil
}

func (uc *AuthUseCase) GetUserFromToken(accessToken string) (*domain.User, error) {
	claims, err := uc.tokenManager.ParseAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}
	userID := claims.UserID

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (uc *AuthUseCase) hashPassword(password string) (string, error) {
	saltedPassword := password + uc.passwordSalt

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (uc *AuthUseCase) comparePasswords(hashedPassword, password string) error {
	saltedPassword := password + uc.passwordSalt

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(saltedPassword))
}
