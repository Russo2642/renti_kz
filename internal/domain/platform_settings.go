package domain

import (
	"time"
)

type PlatformSetting struct {
	ID           int       `json:"id"`
	SettingKey   string    `json:"setting_key"`
	SettingValue string    `json:"setting_value"`
	Description  string    `json:"description"`
	DataType     string    `json:"data_type"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PlatformSettingsRepository interface {
	GetByKey(key string) (*PlatformSetting, error)
	GetAll() ([]*PlatformSetting, error)
	GetAllActive() ([]*PlatformSetting, error)
	Create(setting *PlatformSetting) error
	Update(setting *PlatformSetting) error
	Delete(key string) error
}

type PlatformSettingsUseCase interface {
	GetByKey(key string) (*PlatformSetting, error)
	GetAll() ([]*PlatformSetting, error)
	GetAllActive() ([]*PlatformSetting, error)
	Create(setting *PlatformSetting) error
	Update(setting *PlatformSetting) error
	Delete(key string) error

	GetServiceFeePercentage() (int, error)
	GetMinBookingDurationHours() (int, error)
	GetMaxBookingDurationHours() (int, error)
	GetDefaultCleaningDurationMinutes() (int, error)
	GetPlatformCommissionPercentage() (int, error)
	GetMaxAdvanceBookingDays() (int, error)
}

const (
	SettingKeyServiceFeePercentage           = "service_fee_percentage"
	SettingKeyMinBookingDurationHours        = "min_booking_duration_hours"
	SettingKeyMaxBookingDurationHours        = "max_booking_duration_hours"
	SettingKeyDefaultCleaningDurationMinutes = "default_cleaning_duration_minutes"
	SettingKeyPlatformCommissionPercentage   = "platform_commission_percentage"

	SettingKeyMaxAdvanceBookingDays          = "max_advance_booking_days"
)
