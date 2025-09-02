package domain

import (
	"time"
)

type LockStatus string

const (
	LockStatusClosed LockStatus = "closed"
	LockStatusOpen   LockStatus = "open"
)

type BatteryType string

const (
	BatteryTypeAlkaline BatteryType = "alkaline"
	BatteryTypeLithium  BatteryType = "lithium"
	BatteryTypeUnknown  BatteryType = "unknown"
)

type ChargingStatus string

const (
	ChargingStatusNotCharging ChargingStatus = "not_charging"
	ChargingStatusCharging    ChargingStatus = "charging"
	ChargingStatusFull        ChargingStatus = "full"
)

type LockChangeSource string

const (
	LockChangeSourceAPI     LockChangeSource = "api"
	LockChangeSourceManual  LockChangeSource = "manual"
	LockChangeSourceSystem  LockChangeSource = "system"
	LockChangeSourceTuya    LockChangeSource = "tuya"
	LockChangeSourceWebhook LockChangeSource = "webhook"
)

type Lock struct {
	ID               int        `json:"id"`
	UniqueID         string     `json:"unique_id"`
	ApartmentID      *int       `json:"apartment_id"`
	Apartment        *Apartment `json:"apartment,omitempty"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	CurrentStatus    LockStatus `json:"current_status"`
	LastStatusUpdate *time.Time `json:"last_status_update"`
	LastHeartbeat    *time.Time `json:"last_heartbeat"`
	IsOnline         bool       `json:"is_online"`
	FirmwareVersion  string     `json:"firmware_version"`

	BatteryLevel     *int            `json:"battery_level"`
	BatteryType      BatteryType     `json:"battery_type"`
	ChargingStatus   *ChargingStatus `json:"charging_status"`
	LastBatteryCheck *time.Time      `json:"last_battery_check"`

	SignalStrength *int `json:"signal_strength"`

	TuyaDeviceID      string     `json:"tuya_device_id"`
	LastTuyaSync      *time.Time `json:"last_tuya_sync"`
	AutoUpdateEnabled bool       `json:"auto_update_enabled"`
	WebhookConfigured bool       `json:"webhook_configured"`

	OwnerPassword string    `json:"owner_password"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type LockStatusLog struct {
	ID           int              `json:"id"`
	LockID       int              `json:"lock_id"`
	Lock         *Lock            `json:"lock,omitempty"`
	OldStatus    *LockStatus      `json:"old_status"`
	NewStatus    LockStatus       `json:"new_status"`
	ChangeSource LockChangeSource `json:"change_source"`
	UserID       *int             `json:"user_id"`
	User         *User            `json:"user,omitempty"`
	BookingID    *int             `json:"booking_id"`
	Booking      *Booking         `json:"booking,omitempty"`
	Notes        string           `json:"notes"`
	CreatedAt    time.Time        `json:"created_at"`
}

type LockTempPassword struct {
	ID             int       `json:"id"`
	LockID         int       `json:"lock_id"`
	Lock           *Lock     `json:"lock,omitempty"`
	BookingID      *int      `json:"booking_id"`
	Booking        *Booking  `json:"booking,omitempty"`
	UserID         *int      `json:"user_id"`
	User           *User     `json:"user,omitempty"`
	Password       string    `json:"password"`
	TuyaPasswordID int64     `json:"tuya_password_id"`
	Name           string    `json:"name"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     time.Time `json:"valid_until"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateLockRequest struct {
	UniqueID        string `json:"unique_id" validate:"required"`
	ApartmentID     *int   `json:"apartment_id"`
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description"`
	FirmwareVersion string `json:"firmware_version"`
	TuyaDeviceID    string `json:"tuya_device_id" validate:"required"`
	OwnerPassword   string `json:"owner_password" validate:"required"`
}

type UpdateLockRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	FirmwareVersion string `json:"firmware_version"`
}

type LockStatusUpdateRequest struct {
	UniqueID       string     `json:"unique_id" validate:"required"`
	Status         LockStatus `json:"status" validate:"required"`
	BatteryLevel   *int       `json:"battery_level"`
	SignalStrength *int       `json:"signal_strength"`
	Timestamp      time.Time  `json:"timestamp"`
}

type LockHeartbeatRequest struct {
	UniqueID       string     `json:"unique_id" validate:"required"`
	Status         LockStatus `json:"status" validate:"required"`
	BatteryLevel   *int       `json:"battery_level"`
	SignalStrength *int       `json:"signal_strength"`
	Timestamp      time.Time  `json:"timestamp"`
}
type LockPasswordRequest struct {
	BookingID int `json:"booking_id" validate:"required"`
}

type AdminGeneratePasswordRequest struct {
	Name        string `json:"name" validate:"required"`
	ValidFrom   string `json:"valid_from" validate:"required"`
	ValidUntil  string `json:"valid_until" validate:"required"`
	UserID      *int   `json:"user_id,omitempty"`
	Description string `json:"description,omitempty"`
}

type TuyaWebhookEvent struct {
	BizCode    string                 `json:"bizCode"`
	DevID      string                 `json:"devId"`
	ProductKey string                 `json:"productKey"`
	Ts         int64                  `json:"ts"`
	UUID       string                 `json:"uuid,omitempty"`
	BizData    map[string]interface{} `json:"bizData"`
}

type LockRepository interface {
	Create(lock *Lock) error
	GetByID(id int) (*Lock, error)
	GetByUniqueID(uniqueID string) (*Lock, error)
	GetByApartmentID(apartmentID int) (*Lock, error)
	GetLocksByApartmentID(apartmentID int) ([]*Lock, error)
	GetAll() ([]*Lock, error)
	GetAllWithFilters(filters map[string]interface{}, page, pageSize int) ([]*Lock, int, error)
	Update(lock *Lock) error
	Delete(id int) error

	UpdateStatus(uniqueID string, status LockStatus, timestamp *time.Time) error
	UpdateHeartbeat(uniqueID string, timestamp time.Time, batteryLevel *int, signalStrength *int) error
	UpdateOnlineStatus(uniqueID string, isOnline bool) error

	UpdateBatteryInfo(uniqueID string, batteryLevel *int, batteryType BatteryType, chargingStatus *ChargingStatus) error
	UpdateTuyaSync(uniqueID string, syncTime time.Time) error
	EnableAutoUpdate(uniqueID string, enabled bool) error
	ConfigureWebhook(uniqueID string, configured bool) error
	GetLocksForAutoUpdate() ([]*Lock, error)

	CreateStatusLog(log *LockStatusLog) error
	GetStatusLogsByLockID(lockID int, limit int) ([]*LockStatusLog, error)
	GetStatusLogsByUniqueID(uniqueID string, limit int) ([]*LockStatusLog, error)

	CreateTempPassword(tempPassword *LockTempPassword) error
	GetTempPasswordsByLockID(lockID int) ([]*LockTempPassword, error)
	GetTempPasswordsByBookingID(bookingID int) ([]*LockTempPassword, error)
	GetTempPasswordByID(id int) (*LockTempPassword, error)
	UpdateTempPassword(tempPassword *LockTempPassword) error
	DeleteTempPassword(id int) error
	DeactivateTempPassword(id int) error
}

type LockUseCase interface {
	CreateLock(request *CreateLockRequest) (*Lock, error)
	GetLockByID(id int) (*Lock, error)
	GetLockByUniqueID(uniqueID string) (*Lock, error)
	GetLockByApartmentID(apartmentID int) (*Lock, error)
	GetAllLocks() ([]*Lock, error)
	GetAllWithFilters(filters map[string]interface{}, page, pageSize int) ([]*Lock, int, error)
	UpdateLock(id int, request *UpdateLockRequest) error
	DeleteLock(id int) error

	UpdateLockStatus(request *LockStatusUpdateRequest) error
	ProcessHeartbeat(request *LockHeartbeatRequest) error

	ProcessTuyaWebhookEvent(event *TuyaWebhookEvent) error
	SyncAllLocksWithTuya() error
	SyncLockWithTuya(uniqueID string) error
	EnableAutoUpdate(uniqueID string) error
	DisableAutoUpdate(uniqueID string) error
	ConfigureTuyaWebhooks(uniqueID string) error

	GeneratePasswordForBooking(uniqueID string, userID int, bookingID int) (string, error)
	GeneratePasswordForBookingByID(bookingID, userID int) (string, error)
	GetOwnerPassword(uniqueID string, userID int) (string, error)
	DeactivatePasswordForBooking(bookingID int) error
	ExtendPasswordForBooking(bookingID int, newEndDate time.Time) error

	GetLockStatus(uniqueID string) (*Lock, error)
	GetLockHistory(uniqueID string, limit int) ([]*LockStatusLog, error)

	CanUserControlLock(uniqueID string, userID int) (bool, error)
	CanUserManageLockViaBooking(booking *Booking, userID int) (bool, error)

	CheckOfflineLocks() ([]*Lock, error)

	GetTempPasswordsByLockID(lockID int) ([]*LockTempPassword, error)
	GetTempPasswordsByBookingID(bookingID int) ([]*LockTempPassword, error)

	UpdateOnlineStatus(uniqueID string, isOnline bool) error
	UpdateTuyaSync(uniqueID string, syncTime time.Time) error

	BindLockToApartment(lockID, apartmentID int) error
	UnbindLockFromApartment(lockID int) error
	EmergencyResetLock(lockID int) error

	AdminGeneratePassword(uniqueID string, request *AdminGeneratePasswordRequest) (*LockTempPassword, error)
	AdminGetAllLockPasswords(uniqueID string) ([]*LockTempPassword, error)
	AdminDeactivatePassword(passwordID int) error
}

type TuyaLockService interface {
	GenerateTemporaryPasswordWithTimes(deviceID, name, rawPassword string, validFrom, validUntil time.Time) (string, int64, error)
	DeleteTempPassword(deviceID string, passwordID int64) error
}
