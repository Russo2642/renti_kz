package domain

import (
	"time"
)

type BookingStatus string

const (
	BookingStatusCreated         BookingStatus = "created"
	BookingStatusAwaitingPayment BookingStatus = "awaiting_payment"
	BookingStatusPending         BookingStatus = "pending"
	BookingStatusApproved        BookingStatus = "approved"
	BookingStatusRejected        BookingStatus = "rejected"
	BookingStatusActive          BookingStatus = "active"
	BookingStatusCompleted       BookingStatus = "completed"
	BookingStatusCanceled        BookingStatus = "canceled"
)

type DoorStatus string

const (
	DoorStatusClosed DoorStatus = "closed"
	DoorStatusOpen   DoorStatus = "open"
)

type Booking struct {
	ID                 int           `json:"id"`
	RenterID           int           `json:"renter_id"`
	Renter             *Renter       `json:"renter,omitempty"`
	ApartmentID        int           `json:"apartment_id"`
	Apartment          *Apartment    `json:"apartment,omitempty"`
	ContractID         *int          `json:"contract_id,omitempty"`
	StartDate          time.Time     `json:"start_date"`
	EndDate            time.Time     `json:"end_date"`
	Duration           int           `json:"duration"`
	CleaningDuration   int           `json:"cleaning_duration"`
	Status             BookingStatus `json:"status"`
	TotalPrice         int           `json:"total_price"`
	ServiceFee         int           `json:"service_fee"`
	FinalPrice         int           `json:"final_price"`
	IsContractAccepted bool          `json:"is_contract_accepted"`
	PaymentID          *int64        `json:"payment_id,omitempty"`
	CancellationReason *string       `json:"cancellation_reason"`
	OwnerComment       *string       `json:"owner_comment"`
	BookingNumber      string        `json:"booking_number"`
	DoorStatus         DoorStatus    `json:"door_status"`
	LastDoorAction     *time.Time    `json:"last_door_action"`
	CanExtend          bool          `json:"can_extend"`
	ExtensionRequested bool          `json:"extension_requested"`
	ExtensionEndDate   *time.Time    `json:"extension_end_date"`
	ExtensionDuration  int           `json:"extension_duration"`
	ExtensionPrice     int           `json:"extension_price"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

type BookingExtension struct {
	ID          int           `json:"id"`
	BookingID   int           `json:"booking_id"`
	Booking     *Booking      `json:"booking,omitempty"`
	Duration    int           `json:"duration"`
	Price       int           `json:"price"`
	Status      BookingStatus `json:"status"`
	PaymentID   *int64        `json:"payment_id,omitempty"`
	RequestedAt time.Time     `json:"requested_at"`
	ApprovedAt  *time.Time    `json:"approved_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type DoorAction struct {
	ID        int        `json:"id"`
	BookingID int        `json:"booking_id"`
	Booking   *Booking   `json:"booking,omitempty"`
	UserID    int        `json:"user_id"`
	User      *User      `json:"user,omitempty"`
	Action    DoorStatus `json:"action"`
	Success   bool       `json:"success"`
	Error     string     `json:"error"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateBookingRequest struct {
	ApartmentID int    `json:"apartment_id" validate:"required"`
	StartDate   string `json:"start_date" validate:"required"`
	Duration    int    `json:"duration" validate:"required,min=1"`
}

type ConfirmBookingRequest struct {
	IsContractAccepted bool `json:"is_contract_accepted" validate:"required"`
}

type UpdateBookingStatusRequest struct {
	Status  BookingStatus `json:"status" validate:"required"`
	Comment string        `json:"comment"`
}

type ExtendBookingRequest struct {
	Duration int `json:"duration" validate:"required,min=1"`
}

type AvailableExtensionsResponse struct {
	BookingID               int                    `json:"booking_id"`
	CurrentEndDate          string                 `json:"current_end_date"`
	AvailableExtensions     []int                  `json:"available_extensions"`
	MaxPossibleExtension    int                    `json:"max_possible_extension"`
	NextBookingStartsAt     *string                `json:"next_booking_starts_at,omitempty"`
	CleaningDurationMinutes int                    `json:"cleaning_duration_minutes"`
	Limitations             map[string]string      `json:"limitations,omitempty"`
	TimeInfo                map[string]interface{} `json:"time_info"`
}

type RejectBookingRequest struct {
	Comment string `json:"comment"`
}

type CancelBookingRequest struct {
	Reason string `json:"reason"`
}

type BookingRepository interface {
	Create(booking *Booking) error
	GetByID(id int) (*Booking, error)
	GetByBookingNumber(bookingNumber string) (*Booking, error)
	GetByRenterID(renterID int, status []BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*Booking, int, error)
	GetByApartmentID(apartmentID int, status []BookingStatus) ([]*Booking, error)
	GetByOwnerID(ownerID int, status []BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*Booking, int, error)
	GetByStatus(status []BookingStatus) ([]*Booking, error)
	Update(booking *Booking) error
	UpdateDoorStatus(bookingID int, doorStatus DoorStatus, lastAction *time.Time) error
	Delete(id int) error

	CheckApartmentAvailability(apartmentID int, startDate, endDate time.Time, excludeBookingID *int) (bool, error)
	GetNextBookingAfterDate(apartmentID int, afterDate time.Time, excludeBookingID *int) (*Booking, error)

	CreateExtension(extension *BookingExtension) error
	GetExtensionsByBookingID(bookingID int) ([]*BookingExtension, error)
	GetExtensionByID(extensionID int) (*BookingExtension, error)
	UpdateExtension(extension *BookingExtension) error

	CreateDoorAction(action *DoorAction) error
	GetDoorActionsByBookingID(bookingID int) ([]*DoorAction, error)
	GetLastDoorAction(bookingID int) (*DoorAction, error)

	GetAll(filters map[string]interface{}, page, pageSize int) ([]*Booking, int, error)
	GetStatusStatistics() (map[string]int, error)

	CleanupExpiredBookings(batchSize int) (int, error)
	CleanupExpiredExtensions(batchSize int) (int, error)
}

type BookingUseCase interface {
	CreateBooking(userID int, request *CreateBookingRequest) (*Booking, error)
	ConfirmBooking(bookingID, userID int, request *ConfirmBookingRequest) (*Booking, error)
	ProcessPayment(bookingID int, paymentID string) (*Booking, error)
	ProcessPaymentWithOrder(bookingID int, orderID string) (*Booking, error)
	GetBookingByID(bookingID int) (*Booking, error)
	GetBookingByNumber(bookingNumber string) (*Booking, error)
	GetRenterBookings(userID int, status []BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*Booking, int, error)
	GetOwnerBookings(userID int, status []BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*Booking, int, error)

	ApproveBooking(bookingID, userID int) error
	RejectBooking(bookingID, userID int, comment string) error
	CancelBooking(bookingID, userID int, reason string) error
	FinishSession(bookingID, userID int) error

	RequestExtension(bookingID, userID int, request *ExtendBookingRequest) error
	GetAvailableExtensions(bookingID, userID int) (*AvailableExtensionsResponse, error)
	ProcessExtensionPayment(extensionID int, paymentID string) (*BookingExtension, error)
	ProcessExtensionPaymentWithOrder(extensionID int, orderID string) (*BookingExtension, error)

	ApproveExtension(extensionID, userID int) error
	RejectExtension(extensionID, userID int) error

	GetBookingExtensions(bookingID int) ([]*BookingExtension, error)

	GetAvailableTimeSlots(apartmentID int, date string, duration int) ([]string, error)

	CanUserAccessBooking(bookingID, userID int) (bool, error)
	CanUserManageDoor(bookingID, userID int) (bool, error)
	IsBookingActive(bookingID int) (bool, error)
	CheckApartmentAvailability(apartmentID int, startDate, endDate time.Time) (bool, error)
	CompleteBooking(bookingID int) error

	GetMyBookingsLockAccess(userID int) (*MyBookingsLockAccessResponse, error)
	GetBookingLockAccess(bookingID, userID int) (*BookingLockAccessResponse, error)

	AdminGetAllBookings(filters map[string]interface{}, page, pageSize int) ([]*Booking, int, error)
	AdminGetBookingByID(bookingID int) (*Booking, error)
	AdminUpdateBookingStatus(bookingID int, status BookingStatus, reason string, adminID int) error
	AdminCancelBooking(bookingID int, reason string, adminID int) error
	AdminGetBookingStatistics() (map[string]interface{}, error)

	GetStatusStatistics() (map[string]int, error)
	GetPaymentReceipt(bookingID, userID int) (*PaymentReceipt, error)
}

type BookingResponse struct {
	ID                 int           `json:"id"`
	RenterID           int           `json:"renter_id"`
	Renter             *Renter       `json:"renter,omitempty"`
	ApartmentID        int           `json:"apartment_id"`
	Apartment          *Apartment    `json:"apartment,omitempty"`
	ContractID         *int          `json:"contract_id,omitempty"`
	LockUniqueID       *string       `json:"lock_unique_id,omitempty"`
	StartDate          string        `json:"start_date"`
	EndDate            string        `json:"end_date"`
	Duration           int           `json:"duration"`
	CleaningDuration   int           `json:"cleaning_duration"`
	Status             BookingStatus `json:"status"`
	TotalPrice         int           `json:"total_price"`
	ServiceFee         int           `json:"service_fee"`
	FinalPrice         int           `json:"final_price"`
	IsContractAccepted bool          `json:"is_contract_accepted"`
	CancellationReason *string       `json:"cancellation_reason"`
	OwnerComment       *string       `json:"owner_comment"`
	BookingNumber      string        `json:"booking_number"`
	DoorStatus         DoorStatus    `json:"door_status"`
	LastDoorAction     *string       `json:"last_door_action"`
	CanExtend          bool          `json:"can_extend"`
	ExtensionRequested bool          `json:"extension_requested"`
	ExtensionEndDate   *string       `json:"extension_end_date"`
	ExtensionDuration  int           `json:"extension_duration"`
	ExtensionPrice     int           `json:"extension_price"`
	CreatedAt          string        `json:"created_at"`
	UpdatedAt          string        `json:"updated_at"`
}

type LockAccessStatus string

const (
	LockAccessAvailableNow   LockAccessStatus = "available_now"
	LockAccessAvailableSoon  LockAccessStatus = "available_soon"
	LockAccessPasswordExists LockAccessStatus = "password_exists"
	LockAccessNotAvailable   LockAccessStatus = "not_available"
)

type ApartmentInfo struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Address string `json:"address"`
}

type BookingPeriod struct {
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
	Status    BookingStatus `json:"status"`
}

type LockInfo struct {
	UniqueID string `json:"unique_id"`
	Name     string `json:"name"`
}

type LockAccess struct {
	Status             LockAccessStatus `json:"status"`
	CanGenerateNow     bool             `json:"can_generate_now"`
	PasswordExists     bool             `json:"password_exists"`
	Password           *string          `json:"password,omitempty"`
	PasswordValidFrom  *time.Time       `json:"password_valid_from,omitempty"`
	PasswordValidUntil *time.Time       `json:"password_valid_until,omitempty"`
	AvailableAt        *time.Time       `json:"available_at,omitempty"`
	Message            string           `json:"message"`
	DetailedReason     *string          `json:"detailed_reason,omitempty"`
	UsageInstructions  *string          `json:"usage_instructions,omitempty"`
}

type BookingLockAccess struct {
	BookingID     int           `json:"booking_id"`
	ApartmentInfo ApartmentInfo `json:"apartment_info"`
	BookingPeriod BookingPeriod `json:"booking_period"`
	LockAccess    LockAccess    `json:"lock_access"`
	LockInfo      *LockInfo     `json:"lock_info,omitempty"`
}

type MyBookingsLockAccessResponse struct {
	Bookings []BookingLockAccess `json:"bookings"`
	Total    int                 `json:"total"`
}

type BookingLockAccessResponse struct {
	BookingID  int        `json:"booking_id"`
	LockAccess LockAccess `json:"lock_access"`
}
