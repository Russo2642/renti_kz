package domain

import (
	"encoding/json"
	"time"
)

type ContractType string

const (
	ContractTypeApartment ContractType = "apartment"
	ContractTypeRental    ContractType = "rental"
)

type ContractStatus string

const (
	ContractStatusDraft     ContractStatus = "draft"
	ContractStatusConfirmed ContractStatus = "confirmed"
	ContractStatusSigned    ContractStatus = "signed"
)

type Contract struct {
	ID              int             `json:"id"`
	Type            ContractType    `json:"type"`
	ApartmentID     int             `json:"apartment_id"`
	BookingID       *int            `json:"booking_id,omitempty"`
	TemplateVersion int             `json:"template_version"`
	DataSnapshot    json.RawMessage `json:"data_snapshot,omitempty"`
	Status          ContractStatus  `json:"status"`
	IsActive        bool            `json:"is_active"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`

	Apartment *Apartment `json:"apartment,omitempty"`
	Booking   *Booking   `json:"booking,omitempty"`
}

type RentalContractSnapshot struct {
	BookingNumber string    `json:"booking_number"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Duration      int       `json:"duration"`
	TotalPrice    int       `json:"total_price"`
	ServiceFee    int       `json:"service_fee"`
	FinalPrice    int       `json:"final_price"`

	RenterName string `json:"renter_name"`
	RenterIIN  string `json:"renter_iin"`
	OwnerName  string `json:"owner_name"`
	OwnerIIN   string `json:"owner_iin"`

	ApartmentAddress string `json:"apartment_address"`
	ApartmentTitle   string `json:"apartment_title"`

	ContractDate    string `json:"contract_date"`
	CreatedByUserID int    `json:"created_by_user_id"`
}

type ApartmentContractSnapshot struct {
	ContractDate    string `json:"contract_date"`
	CreatedByUserID int    `json:"created_by_user_id"`
}

type ContractContactInfo struct {
	RenterPhone string `json:"renter_phone"`
	RenterEmail string `json:"renter_email"`
	OwnerPhone  string `json:"owner_phone"`
	OwnerEmail  string `json:"owner_email"`
}

type ContractTemplateData struct {
	*RentalContractSnapshot
	*ApartmentContractSnapshot

	*ContractContactInfo

	ContractDate    string `json:"contract_date"`
	TemplateVersion int    `json:"template_version"`
}

type ContractTemplate struct {
	ID              int          `json:"id"`
	Type            ContractType `json:"type"`
	Version         int          `json:"version"`
	TemplateContent string       `json:"template_content"`
	IsActive        bool         `json:"is_active"`
	CreatedAt       time.Time    `json:"created_at"`
}

type ContractRepository interface {
	Create(contract *Contract) error
	GetByID(id int) (*Contract, error)
	Update(contract *Contract) error
	Delete(id int) error

	GetByBookingID(bookingID int) (*Contract, error)
	GetByApartmentID(apartmentID int, contractType ContractType) (*Contract, error)
	GetByApartmentIDAndType(apartmentID int, contractType ContractType) ([]*Contract, error)
	GetActiveByApartmentID(apartmentID int) ([]*Contract, error)

	GetByStatus(status ContractStatus, limit, offset int) ([]*Contract, error)
	GetByType(contractType ContractType, limit, offset int) ([]*Contract, error)
	GetAll(limit, offset int) ([]*Contract, error)

	ExistsForBooking(bookingID int) (bool, error)
	GetContractIDByBookingID(bookingID int) (*int, error)
	CountByType(contractType ContractType) (int, error)
}

type ContractTemplateRepository interface {
	GetByTypeAndVersion(contractType ContractType, version int) (*ContractTemplate, error)
	GetActiveByType(contractType ContractType) (*ContractTemplate, error)
	GetLatestVersion(contractType ContractType) (int, error)
	Create(template *ContractTemplate) error
	Update(template *ContractTemplate) error
	SetActiveVersion(contractType ContractType, version int) error
}

type ContractService interface {
	GenerateContractHTML(contractID int) (string, error)
	GetOrGenerateContractHTML(contractID int) (string, error)

	InvalidateContractCache(contractID int) error
	WarmupContractCache(contractID int) error

	RenderTemplate(contractType ContractType, data *ContractTemplateData) (string, error)
}

type ContractUseCase interface {
	CreateRentalContract(bookingID int) (*Contract, error)
	CreateApartmentContract(apartmentID int) (*Contract, error)

	GetContractByID(id int) (*Contract, error)
	GetContractByBookingID(bookingID int) (*Contract, error)
	GetApartmentContract(apartmentID int) (*Contract, error)
	GetContractHTML(contractID int) (string, error)

	UpdateContractStatus(contractID int, status ContractStatus) error
	ConfirmContract(contractID int) error

	CanUserAccessContract(contractID, userID int) (bool, error)

	GetUserRentalContracts(userID int) ([]*Contract, error)
	GetOwnerApartmentContracts(ownerID int) ([]*Contract, error)
	GetAllContracts(limit, offset int) ([]*Contract, error)

	RefreshContractData(contractID int) error
}

type ContractResponse struct {
	ID              int            `json:"id"`
	Type            ContractType   `json:"type"`
	ApartmentID     int            `json:"apartment_id"`
	BookingID       *int           `json:"booking_id,omitempty"`
	TemplateVersion int            `json:"template_version"`
	Status          ContractStatus `json:"status"`
	IsActive        bool           `json:"is_active"`
	ExpiresAt       *string        `json:"expires_at,omitempty"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`

	Apartment *Apartment `json:"apartment,omitempty"`
	Booking   *Booking   `json:"booking,omitempty"`
}

type ContractHTMLResponse struct {
	ContractID int            `json:"contract_id"`
	HTML       string         `json:"html"`
	Status     ContractStatus `json:"status"`
	CachedAt   string         `json:"cached_at"`
}

func (c *Contract) IsRentalContract() bool {
	return c.Type == ContractTypeRental
}

func (c *Contract) IsApartmentContract() bool {
	return c.Type == ContractTypeApartment
}

func (c *Contract) IsExpired() bool {
	return c.ExpiresAt != nil && time.Now().After(*c.ExpiresAt)
}

func (c *Contract) GetSnapshotData() (*RentalContractSnapshot, error) {
	if c.DataSnapshot == nil {
		return nil, nil
	}

	var snapshot RentalContractSnapshot
	err := json.Unmarshal(c.DataSnapshot, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (c *Contract) GetApartmentSnapshotData() (*ApartmentContractSnapshot, error) {
	if c.DataSnapshot == nil {
		return nil, nil
	}

	var snapshot ApartmentContractSnapshot
	err := json.Unmarshal(c.DataSnapshot, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (c *Contract) SetSnapshotData(snapshot interface{}) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	c.DataSnapshot = data
	return nil
}
