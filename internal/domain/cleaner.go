package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Cleaner struct {
	ID         int              `json:"id"`
	UserID     int              `json:"user_id"`
	User       *User            `json:"user,omitempty"`
	Apartments []*Apartment     `json:"apartments,omitempty"`
	IsActive   bool             `json:"is_active"`
	Schedule   *CleanerSchedule `json:"schedule,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type CleanerApartment struct {
	ID          int        `json:"id"`
	CleanerID   int        `json:"cleaner_id"`
	ApartmentID int        `json:"apartment_id"`
	Apartment   *Apartment `json:"apartment,omitempty"`
	IsActive    bool       `json:"is_active"`
	AssignedAt  time.Time  `json:"assigned_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CleanerSchedule struct {
	Monday    []WorkingHours `json:"monday,omitempty"`
	Tuesday   []WorkingHours `json:"tuesday,omitempty"`
	Wednesday []WorkingHours `json:"wednesday,omitempty"`
	Thursday  []WorkingHours `json:"thursday,omitempty"`
	Friday    []WorkingHours `json:"friday,omitempty"`
	Saturday  []WorkingHours `json:"saturday,omitempty"`
	Sunday    []WorkingHours `json:"sunday,omitempty"`
}

type CleanerSchedulePatch struct {
	Monday    *[]WorkingHours `json:"monday,omitempty"`
	Tuesday   *[]WorkingHours `json:"tuesday,omitempty"`
	Wednesday *[]WorkingHours `json:"wednesday,omitempty"`
	Thursday  *[]WorkingHours `json:"thursday,omitempty"`
	Friday    *[]WorkingHours `json:"friday,omitempty"`
	Saturday  *[]WorkingHours `json:"saturday,omitempty"`
	Sunday    *[]WorkingHours `json:"sunday,omitempty"`
}

func (cs *CleanerSchedule) Value() (driver.Value, error) {
	if cs == nil {
		return nil, nil
	}
	return json.Marshal(cs)
}

func (cs *CleanerSchedule) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan into CleanerSchedule")
	}

	return json.Unmarshal(bytes, cs)
}

type ApartmentForCleaning struct {
	ID                   int        `json:"id"`
	Street               string     `json:"street"`
	Building             string     `json:"building"`
	ApartmentNumber      int        `json:"apartment_number"`
	RoomCount            int        `json:"room_count"`
	TotalArea            float64    `json:"total_area"`
	IsFree               bool       `json:"is_free"`
	LastBookingEndDate   *time.Time `json:"last_booking_end_date,omitempty"`
	TimeSinceLastBooking *string    `json:"time_since_last_booking,omitempty"`
	CleaningStatus       string     `json:"cleaning_status"` // "needs_cleaning", "in_progress", "completed"
	CleaningNotes        *string    `json:"cleaning_notes,omitempty"`
	OwnerContact         *string    `json:"owner_contact,omitempty"`
	City                 *City      `json:"city,omitempty"`
	District             *District  `json:"district,omitempty"`
}

type CreateCleanerRequest struct {
	UserID       int              `json:"user_id" validate:"required"`
	ApartmentIDs []int            `json:"apartment_ids" validate:"required,min=1"`
	Schedule     *CleanerSchedule `json:"schedule,omitempty"`
}

type AssignCleanerToApartmentRequest struct {
	CleanerID   int `json:"cleaner_id" validate:"required"`
	ApartmentID int `json:"apartment_id" validate:"required"`
}

type RemoveCleanerFromApartmentRequest struct {
	CleanerID   int `json:"cleaner_id" validate:"required"`
	ApartmentID int `json:"apartment_id" validate:"required"`
}

type UpdateCleanerRequest struct {
	IsActive *bool            `json:"is_active,omitempty"`
	Schedule *CleanerSchedule `json:"schedule,omitempty"`
}

type UpdateCleanerScheduleRequest struct {
	Schedule *CleanerSchedule `json:"schedule" validate:"required"`
}

type StartCleaningRequest struct {
	ApartmentID int    `json:"apartment_id" validate:"required"`
	Notes       string `json:"notes,omitempty"`
}

type CompleteCleaningRequest struct {
	ApartmentID  int      `json:"apartment_id" validate:"required"`
	Notes        string   `json:"notes,omitempty"`
	PhotosBase64 []string `json:"photos_base64,omitempty"`
}

type CleanerRepository interface {
	Create(cleaner *Cleaner) error
	GetByID(id int) (*Cleaner, error)
	GetByUserID(userID int) (*Cleaner, error)
	GetByApartmentID(apartmentID int) ([]*Cleaner, error)
	GetByApartmentIDActive(apartmentID int) ([]*Cleaner, error)
	GetAll(filters map[string]interface{}, page, pageSize int) ([]*Cleaner, int, error)
	Update(cleaner *Cleaner) error
	UpdatePartial(id int, request *UpdateCleanerRequest) error
	Delete(id int) error
	IsUserCleaner(userID int) (bool, error)
	GetCleanersByOwner(ownerID int) ([]*Cleaner, error)
	UpdateSchedulePatch(cleanerID int, schedulePatch *CleanerSchedulePatch) error

	AssignToApartment(cleanerID, apartmentID int) error
	RemoveFromApartment(cleanerID, apartmentID int) error
	GetCleanerApartments(cleanerID int) ([]*CleanerApartment, error)
	GetApartmentCleaners(apartmentID int) ([]*CleanerApartment, error)

	GetApartmentsForCleaning(cleanerID int) ([]*ApartmentForCleaning, error)
	GetApartmentsNeedingCleaning() ([]*ApartmentForCleaning, error)
}

type CleanerUseCase interface {
	CreateCleaner(request *CreateCleanerRequest, adminID int) (*Cleaner, error)
	GetCleanerByID(id int) (*Cleaner, error)
	GetCleanersByApartment(apartmentID int) ([]*Cleaner, error)
	GetAllCleaners(filters map[string]interface{}, page, pageSize int) ([]*Cleaner, int, error)
	UpdateCleaner(id int, request *UpdateCleanerRequest, adminID int) error
	DeleteCleaner(id int, adminID int) error
	GetCleanersByOwner(ownerID int) ([]*Cleaner, error)
	IsUserCleaner(userID int) (bool, error)
	AssignCleanerToApartment(cleanerID, apartmentID, adminID int) error
	RemoveCleanerFromApartment(cleanerID, apartmentID, adminID int) error

	GetCleanerByUserID(userID int) (*Cleaner, error)
	GetCleanerApartments(userID int) ([]*Apartment, error)
	GetApartmentsForCleaning(userID int) ([]*ApartmentForCleaning, error)
	GetCleanerStats(userID int) (map[string]interface{}, error)
	UpdateCleanerSchedule(userID int, schedule *CleanerSchedule) error
	UpdateCleanerSchedulePatch(userID int, schedule *CleanerSchedulePatch) error

	StartCleaning(userID int, request *StartCleaningRequest) error
	CompleteCleaning(userID int, request *CompleteCleaningRequest) error
	GetCleaningHistory(userID int, filters map[string]interface{}, page, pageSize int) ([]map[string]interface{}, int, error)
	GetApartmentsNeedingCleaning() ([]*ApartmentForCleaning, error)
}

type CleanerResponse struct {
	ID         int              `json:"id"`
	UserID     int              `json:"user_id"`
	User       *User            `json:"user,omitempty"`
	Apartments []*Apartment     `json:"apartments,omitempty"`
	IsActive   bool             `json:"is_active"`
	Schedule   *CleanerSchedule `json:"schedule,omitempty"`
	CreatedAt  string           `json:"created_at"`
	UpdatedAt  string           `json:"updated_at"`
}
