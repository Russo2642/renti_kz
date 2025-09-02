package usecase

import (
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/auth"
)

type OTPUseCase struct {
	otpService   domain.OTPService
	otpRepo      domain.OTPRepository
	userRepo     domain.UserRepository
	tokenManager auth.TokenManager
}

func NewOTPUseCase(
	otpService domain.OTPService,
	otpRepo domain.OTPRepository,
	userRepo domain.UserRepository,
	tokenManager auth.TokenManager,
) *OTPUseCase {
	return &OTPUseCase{
		otpService:   otpService,
		otpRepo:      otpRepo,
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (uc *OTPUseCase) RequestOTP(phone string) (*domain.OTPRequestResponse, error) {
	existingSession, err := uc.otpRepo.GetSessionByPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существующей OTP сессии: %w", err)
	}

	if existingSession != nil {
		err = uc.otpRepo.DeleteSession(existingSession.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка удаления старой OTP сессии: %w", err)
		}
	}

	otpResponse, err := uc.otpService.RequestOTP(phone)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки OTP: %w", err)
	}

	session := &domain.OTPSession{
		ID:         otpResponse.ID,
		Phone:      phone,
		Code:       "",
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		IsVerified: false,
		IsUsed:     false,
		CreatedAt:  time.Now(),
	}

	err = uc.otpRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания OTP сессии: %w", err)
	}

	return otpResponse, nil
}

func (uc *OTPUseCase) VerifyOTP(phone, code string) (bool, error) {
	session, err := uc.otpRepo.GetSessionByPhone(phone)
	if err != nil {
		return false, fmt.Errorf("ошибка получения OTP сессии: %w", err)
	}

	if session == nil {
		return false, fmt.Errorf("OTP сессия не найдена")
	}

	if time.Now().After(session.ExpiresAt) {
		uc.otpRepo.DeleteSession(session.ID)
		return false, fmt.Errorf("время действия OTP кода истекло")
	}

	if session.IsUsed {
		return false, fmt.Errorf("OTP код уже был использован")
	}

	verifyResponse, err := uc.otpService.VerifyOTP(session.ID, code)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки OTP: %w", err)
	}

	if !verifyResponse.Validated {
		return false, nil
	}

	session.IsVerified = true
	session.IsUsed = true
	err = uc.otpRepo.UpdateSession(session)
	if err != nil {
		return false, fmt.Errorf("ошибка обновления OTP сессии: %w", err)
	}

	return true, nil
}

func (uc *OTPUseCase) VerifyOTPAndAuthenticate(id, phone, code string) (*domain.OTPAuthResponse, error) {
	session, err := uc.otpRepo.GetSessionByID(id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения OTP сессии: %w", err)
	}

	if session == nil {
		return &domain.OTPAuthResponse{
			RequiresRegistration: false,
			Message:              "Сессия OTP не найдена или истекла",
			ErrorType:            domain.OTPErrorNotFound,
		}, nil
	}

	if session.Phone != phone {
		return &domain.OTPAuthResponse{
			RequiresRegistration: false,
			Message:              "Номер телефона не совпадает с сессией",
			ErrorType:            domain.OTPErrorPhoneMismatch,
		}, nil
	}

	if time.Now().After(session.ExpiresAt) {
		uc.otpRepo.DeleteSession(id)
		return &domain.OTPAuthResponse{
			RequiresRegistration: false,
			Message:              "Время действия OTP кода истекло",
			ErrorType:            domain.OTPErrorExpired,
		}, nil
	}

	if session.IsUsed {
		return &domain.OTPAuthResponse{
			RequiresRegistration: false,
			Message:              "OTP код уже был использован",
			ErrorType:            domain.OTPErrorAlreadyUsed,
		}, nil
	}

	verifyResponse, err := uc.otpService.VerifyOTP(id, code)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки OTP: %w", err)
	}

	if !verifyResponse.Validated {
		return &domain.OTPAuthResponse{
			RequiresRegistration: false,
			Message:              "Неверный OTP код",
			ErrorType:            domain.OTPErrorInvalidCode,
		}, nil
	}

	session.IsVerified = true
	session.IsUsed = true

	err = uc.otpRepo.UpdateSession(session)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления OTP сессии: %w", err)
	}

	user, err := uc.userRepo.GetByPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	if user == nil {
		uc.otpRepo.DeleteSession(id)

		return &domain.OTPAuthResponse{
			RequiresRegistration: true,
			Message:              "Пользователь не найден. Необходима регистрация.",
		}, nil
	}

	accessToken, refreshToken, err := uc.tokenManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	uc.otpRepo.DeleteSession(id)

	return &domain.OTPAuthResponse{
		RequiresRegistration: false,
		AccessToken:          &accessToken,
		RefreshToken:         &refreshToken,
		User:                 user,
		Message:              "Аутентификация успешна",
	}, nil
}

func (uc *OTPUseCase) CheckStatus(id string) (*domain.OTPStatusResponse, error) {
	session, err := uc.otpRepo.GetSessionByID(id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения локальной OTP сессии: %w", err)
	}

	if session == nil {
		return &domain.OTPStatusResponse{
			Phone:          "",
			Validated:      false,
			ValidationDate: nil,
		}, nil
	}

	if session.IsVerified {
		validationTime := session.CreatedAt.Unix()
		return &domain.OTPStatusResponse{
			Phone:          session.Phone,
			Validated:      true,
			ValidationDate: &validationTime,
		}, nil
	}

	statusResponse, err := uc.otpService.CheckStatus(id)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки статуса OTP: %w", err)
	}

	return statusResponse, nil
}
