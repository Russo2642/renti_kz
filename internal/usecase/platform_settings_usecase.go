package usecase

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/russo2642/renti_kz/internal/domain"
)

type platformSettingsUseCase struct {
	settingsRepo domain.PlatformSettingsRepository
}

func NewPlatformSettingsUseCase(settingsRepo domain.PlatformSettingsRepository) domain.PlatformSettingsUseCase {
	return &platformSettingsUseCase{
		settingsRepo: settingsRepo,
	}
}

func (u *platformSettingsUseCase) GetByKey(key string) (*domain.PlatformSetting, error) {
	return u.settingsRepo.GetByKey(key)
}

func (u *platformSettingsUseCase) GetAll() ([]*domain.PlatformSetting, error) {
	return u.settingsRepo.GetAll()
}

func (u *platformSettingsUseCase) GetAllActive() ([]*domain.PlatformSetting, error) {
	return u.settingsRepo.GetAllActive()
}

func (u *platformSettingsUseCase) Create(setting *domain.PlatformSetting) error {
	if err := u.validateSetting(setting); err != nil {
		return err
	}

	return u.settingsRepo.Create(setting)
}

func (u *platformSettingsUseCase) Update(setting *domain.PlatformSetting) error {
	if err := u.validateSetting(setting); err != nil {
		return err
	}

	return u.settingsRepo.Update(setting)
}

func (u *platformSettingsUseCase) Delete(key string) error {
	return u.settingsRepo.Delete(key)
}

func (u *platformSettingsUseCase) GetServiceFeePercentage() (int, error) {
	setting, err := u.settingsRepo.GetByKey(domain.SettingKeyServiceFeePercentage)
	if err != nil {
		return 15, nil
	}

	value, err := strconv.Atoi(setting.SettingValue)
	if err != nil {
		return 15, fmt.Errorf("некорректное значение сервисного сбора: %w", err)
	}

	if value < 0 || value > 100 {
		return 15, fmt.Errorf("сервисный сбор должен быть от 0 до 100 процентов")
	}

	return value, nil
}

func (u *platformSettingsUseCase) GetMinBookingDurationHours() (int, error) {
	setting, err := u.settingsRepo.GetByKey(domain.SettingKeyMinBookingDurationHours)
	if err != nil {
		return 1, nil
	}

	value, err := strconv.Atoi(setting.SettingValue)
	if err != nil {
		return 1, fmt.Errorf("некорректное значение минимальной продолжительности: %w", err)
	}

	return value, nil
}

func (u *platformSettingsUseCase) GetMaxBookingDurationHours() (int, error) {
	setting, err := u.settingsRepo.GetByKey(domain.SettingKeyMaxBookingDurationHours)
	if err != nil {
		return 720, nil
	}

	value, err := strconv.Atoi(setting.SettingValue)
	if err != nil {
		return 720, fmt.Errorf("некорректное значение максимальной продолжительности: %w", err)
	}

	return value, nil
}

func (u *platformSettingsUseCase) GetDefaultCleaningDurationMinutes() (int, error) {
	setting, err := u.settingsRepo.GetByKey(domain.SettingKeyDefaultCleaningDurationMinutes)
	if err != nil {
		return 60, nil
	}

	value, err := strconv.Atoi(setting.SettingValue)
	if err != nil {
		return 60, fmt.Errorf("некорректное значение времени уборки: %w", err)
	}

	return value, nil
}

func (u *platformSettingsUseCase) GetPlatformCommissionPercentage() (int, error) {
	setting, err := u.settingsRepo.GetByKey(domain.SettingKeyPlatformCommissionPercentage)
	if err != nil {
		return 5, nil
	}

	value, err := strconv.Atoi(setting.SettingValue)
	if err != nil {
		return 5, fmt.Errorf("некорректное значение комиссии платформы: %w", err)
	}

	return value, nil
}

func (u *platformSettingsUseCase) GetMaxAdvanceBookingDays() (int, error) {
	setting, err := u.settingsRepo.GetByKey(domain.SettingKeyMaxAdvanceBookingDays)
	if err != nil {
		return 90, nil
	}

	value, err := strconv.Atoi(setting.SettingValue)
	if err != nil {
		return 90, fmt.Errorf("некорректное значение максимального периода бронирования: %w", err)
	}

	return value, nil
}

func (u *platformSettingsUseCase) validateSetting(setting *domain.PlatformSetting) error {
	if setting.SettingKey == "" {
		return fmt.Errorf("ключ настройки не может быть пустым")
	}

	if setting.SettingValue == "" {
		return fmt.Errorf("значение настройки не может быть пустым")
	}

	switch setting.DataType {
	case "integer":
		if _, err := strconv.Atoi(setting.SettingValue); err != nil {
			return fmt.Errorf("значение должно быть целым числом: %w", err)
		}
	case "decimal":
		if _, err := strconv.ParseFloat(setting.SettingValue, 64); err != nil {
			return fmt.Errorf("значение должно быть числом: %w", err)
		}
	case "boolean":
		value := strings.ToLower(setting.SettingValue)
		if value != "true" && value != "false" && value != "1" && value != "0" {
			return fmt.Errorf("значение должно быть true/false или 1/0")
		}
	}

	switch setting.SettingKey {
	case domain.SettingKeyServiceFeePercentage:
		value, _ := strconv.Atoi(setting.SettingValue)
		if value < 0 || value > 100 {
			return fmt.Errorf("сервисный сбор должен быть от 0 до 100 процентов")
		}
	case domain.SettingKeyPlatformCommissionPercentage:
		value, _ := strconv.Atoi(setting.SettingValue)
		if value < 0 || value > 50 {
			return fmt.Errorf("комиссия платформы должна быть от 0 до 50 процентов")
		}
	}

	return nil
}
