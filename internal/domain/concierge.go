package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Concierge struct {
	ID         int                `json:"id"`
	UserID     int                `json:"user_id"`
	User       *User              `json:"user,omitempty"`
	Apartments []*Apartment       `json:"apartments,omitempty"`
	IsActive   bool               `json:"is_active"`
	Schedule   *ConciergeSchedule `json:"schedule,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type ConciergeApartment struct {
	ID          int        `json:"id"`
	ConciergeID int        `json:"concierge_id"`
	ApartmentID int        `json:"apartment_id"`
	Apartment   *Apartment `json:"apartment,omitempty"`
	IsActive    bool       `json:"is_active"`
	AssignedAt  time.Time  `json:"assigned_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ConciergeSchedule struct {
	Monday    []WorkingHours `json:"monday"`
	Tuesday   []WorkingHours `json:"tuesday"`
	Wednesday []WorkingHours `json:"wednesday"`
	Thursday  []WorkingHours `json:"thursday"`
	Friday    []WorkingHours `json:"friday"`
	Saturday  []WorkingHours `json:"saturday"`
	Sunday    []WorkingHours `json:"sunday"`
}

func (cs *ConciergeSchedule) Value() (driver.Value, error) {
	if cs == nil {
		return nil, nil
	}
	return json.Marshal(cs)
}

func (cs *ConciergeSchedule) Scan(value interface{}) error {
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
		return errors.New("cannot scan into ConciergeSchedule")
	}

	return json.Unmarshal(bytes, cs)
}

type WorkingHours struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type CreateConciergeRequest struct {
	UserID       int                `json:"user_id" validate:"required"`
	ApartmentIDs []int              `json:"apartment_ids" validate:"required,min=1"`
	Schedule     *ConciergeSchedule `json:"schedule,omitempty"`
}

type AssignConciergeToApartmentRequest struct {
	ConciergeID int `json:"concierge_id" validate:"required"`
	ApartmentID int `json:"apartment_id" validate:"required"`
}

type RemoveConciergeFromApartmentRequest struct {
	ConciergeID int `json:"concierge_id" validate:"required"`
	ApartmentID int `json:"apartment_id" validate:"required"`
}

type UpdateConciergeRequest struct {
	IsActive bool               `json:"is_active"`
	Schedule *ConciergeSchedule `json:"schedule,omitempty"`
}

type ConciergeRepository interface {
	Create(concierge *Concierge) error
	GetByID(id int) (*Concierge, error)
	GetByUserID(userID int) (*Concierge, error)
	GetByApartmentID(apartmentID int) ([]*Concierge, error)
	GetByApartmentIDActive(apartmentID int) ([]*Concierge, error)
	GetAll(filters map[string]interface{}, page, pageSize int) ([]*Concierge, int, error)
	Update(concierge *Concierge) error
	Delete(id int) error
	IsUserConcierge(userID int) (bool, error)
	GetConciergesByOwner(ownerID int) ([]*Concierge, error)

	AssignToApartment(conciergeID, apartmentID int) error
	RemoveFromApartment(conciergeID, apartmentID int) error
	GetConciergeApartments(conciergeID int) ([]*ConciergeApartment, error)
	GetApartmentConcierges(apartmentID int) ([]*ConciergeApartment, error)
}

type ConciergeUseCase interface {
	CreateConcierge(request *CreateConciergeRequest, adminID int) (*Concierge, error)
	GetConciergeByID(id int) (*Concierge, error)
	GetConciergesByApartment(apartmentID int) ([]*Concierge, error)
	GetAllConcierges(filters map[string]interface{}, page, pageSize int) ([]*Concierge, int, error)
	UpdateConcierge(id int, request *UpdateConciergeRequest, adminID int) error
	DeleteConcierge(id int, adminID int) error
	GetConciergesByOwner(ownerID int) ([]*Concierge, error)
	IsUserConcierge(userID int) (bool, error)
	AssignConciergeToApartment(conciergeID, apartmentID, adminID int) error
	RemoveConciergeFromApartment(conciergeID, apartmentID, adminID int) error

	GetConciergeByUserID(userID int) (*Concierge, error)
	GetConciergeApartments(userID int) ([]*Apartment, error)
	GetConciergeBookings(userID int, filters map[string]interface{}, page, pageSize int) ([]*Booking, int, error)
	GetConciergeStats(userID int) (map[string]interface{}, error)
	UpdateConciergeSchedule(userID int, schedule *ConciergeSchedule) error
}

type ConciergeResponse struct {
	ID         int                `json:"id"`
	UserID     int                `json:"user_id"`
	User       *User              `json:"user,omitempty"`
	Apartments []*Apartment       `json:"apartments,omitempty"`
	IsActive   bool               `json:"is_active"`
	Schedule   *ConciergeSchedule `json:"schedule,omitempty"`
	CreatedAt  string             `json:"created_at"`
	UpdatedAt  string             `json:"updated_at"`
}
