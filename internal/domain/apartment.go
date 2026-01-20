package domain

import (
	"time"
)

const (
	DocumentTypeOwner    = "owner"
	DocumentTypeRealtor  = "realtor"
	ServiceFeePercentage = 15

	ListingTypeOwner   = "owner"
	ListingTypeRealtor = "realtor"
)

type ApartmentStatus string

const (
	AptStatusPending       ApartmentStatus = "pending"
	AptStatusApproved      ApartmentStatus = "approved"
	AptStatusNeedsRevision ApartmentStatus = "needs_revision"
	AptStatusRejected      ApartmentStatus = "rejected"
)

type ApartmentCondition struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ApartmentLocation struct {
	ID          int        `json:"id"`
	ApartmentID int        `json:"apartment_id"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Apartment   *Apartment `json:"apartment,omitempty"`
}

type ApartmentCoordinates struct {
	ID        int     `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ApartmentPhoto struct {
	ID          int        `json:"id"`
	ApartmentID int        `json:"apartment_id"`
	URL         string     `json:"url"`
	Order       int        `json:"order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Apartment   *Apartment `json:"apartment,omitempty"`
}

type ApartmentDocument struct {
	ID          int        `json:"id"`
	ApartmentID int        `json:"apartment_id"`
	URL         string     `json:"url"`
	Type        string     `json:"type"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Apartment   *Apartment `json:"apartment,omitempty"`
}

type Apartment struct {
	ID                   int                  `json:"id"`
	OwnerID              int                  `json:"owner_id"`
	Owner                *PropertyOwner       `json:"owner,omitempty"`
	CityID               int                  `json:"city_id"`
	City                 *City                `json:"city,omitempty"`
	DistrictID           int                  `json:"district_id"`
	District             *District            `json:"district,omitempty"`
	MicrodistrictID      *int                 `json:"microdistrict_id,omitempty"`
	Microdistrict        *Microdistrict       `json:"microdistrict,omitempty"`
	ApartmentTypeID      *int                 `json:"apartment_type_id,omitempty"`
	ApartmentType        *ApartmentType       `json:"apartment_type,omitempty"`
	Street               string               `json:"street"`
	Building             string               `json:"building"`
	ApartmentNumber      int                  `json:"apartment_number"`
	ResidentialComplex   *string              `json:"residential_complex,omitempty"`
	RoomCount            int                  `json:"room_count"`
	TotalArea            float64              `json:"total_area"`
	KitchenArea          float64              `json:"kitchen_area"`
	Floor                int                  `json:"floor"`
	TotalFloors          int                  `json:"total_floors"`
	ConditionID          int                  `json:"condition_id"`
	Condition            *ApartmentCondition  `json:"condition,omitempty"`
	Price                int                  `json:"price"`
	DailyPrice           int                  `json:"daily_price"`
	ServiceFeePercentage int                  `json:"service_fee_percentage"`
	RentalTypeHourly     bool                 `json:"rental_type_hourly"`
	RentalTypeDaily      bool                 `json:"rental_type_daily"`
	IsFree               bool                 `json:"is_free"`
	IsFavorite           bool                 `json:"is_favorite"`
	Status               ApartmentStatus      `json:"status"`
	ModeratorComment     string               `json:"moderator_comment,omitempty"`
	Description          string               `json:"description"`
	ListingType          string               `json:"listing_type"`
	IsAgreementAccepted  bool                 `json:"is_agreement_accepted"`
	AgreementAcceptedAt  *time.Time           `json:"agreement_accepted_at,omitempty"`
	ContractID           *int                 `json:"contract_id,omitempty"`
	HouseRules           []*HouseRules        `json:"house_rules,omitempty"`
	Amenities            []*PopularAmenities  `json:"amenities,omitempty"`
	Photos               []*ApartmentPhoto    `json:"photos,omitempty"`
	Documents            []*ApartmentDocument `json:"documents,omitempty"`
	Location             *ApartmentLocation   `json:"location,omitempty"`
	ViewCount            int                  `json:"view_count"`
	BookingCount         int                  `json:"booking_count"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

type ApartmentRepository interface {
	Create(apartment *Apartment) error
	GetByID(id int) (*Apartment, error)
	GetByIDWithUserContext(id int, userID *int) (*Apartment, error)
	GetByOwnerID(ownerID int) ([]*Apartment, error)
	GetAll(filters map[string]interface{}, page, pageSize int) ([]*Apartment, int, error)
	GetAllWithUserContext(filters map[string]interface{}, page, pageSize int, userID *int) ([]*Apartment, int, error)
	GetByCoordinates(minLat, maxLat, minLng, maxLng float64) ([]*ApartmentCoordinates, error)
	GetByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*ApartmentCoordinates, error)
	GetFullApartmentsByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*Apartment, error)
	GetStatusStatistics() (map[string]int, error)
	GetCityStatistics() (map[string]int, error)
	GetDistrictStatistics() (map[string]int, error)
	Update(apartment *Apartment) error
	UpdateIsFree(apartmentID int, isFree bool) error
	UpdateMultipleIsFree(apartmentStatusMap map[int]bool) error
	Delete(id int) error
	IncrementViewCount(apartmentID int) error
	IncrementBookingCount(apartmentID int) error
	AdminUpdateViewCount(apartmentID int, viewCount int) error
	AdminUpdateBookingCount(apartmentID int, bookingCount int) error
	AdminResetCounters(apartmentID int) error

	AddPhoto(photo *ApartmentPhoto) error
	GetPhotosByApartmentID(apartmentID int) ([]*ApartmentPhoto, error)
	GetPhotoByID(id int) (*ApartmentPhoto, error)
	DeletePhoto(id int) error

	AddDocument(document *ApartmentDocument) error
	GetDocumentsByApartmentID(apartmentID int) ([]*ApartmentDocument, error)
	GetDocumentsByApartmentIDAndType(apartmentID int, documentType string) ([]*ApartmentDocument, error)
	GetDocumentByID(id int) (*ApartmentDocument, error)
	DeleteDocument(id int) error

	AddLocation(location *ApartmentLocation) error
	UpdateLocation(location *ApartmentLocation) error
	GetLocationByApartmentID(apartmentID int) (*ApartmentLocation, error)

	GetAllConditions() ([]*ApartmentCondition, error)
	GetConditionByID(id int) (*ApartmentCondition, error)
	CreateCondition(condition *ApartmentCondition) error
	UpdateCondition(condition *ApartmentCondition) error

	GetHouseRulesByID(id int) (*HouseRules, error)
	GetAllHouseRules() ([]*HouseRules, error)
	GetPopularAmenitiesByID(id int) (*PopularAmenities, error)
	GetAllPopularAmenities() ([]*PopularAmenities, error)

	AddHouseRulesToApartment(apartmentID int, houseRuleIDs []int) error
	AddAmenitiesToApartment(apartmentID int, amenityIDs []int) error
	GetHouseRulesByApartmentID(apartmentID int) ([]*HouseRules, error)
	GetAmenitiesByApartmentID(apartmentID int) ([]*PopularAmenities, error)
}

type ApartmentUseCase interface {
	Create(apartment *Apartment) error
	GetByID(id int) (*Apartment, error)
	GetByIDWithUserContext(id int, userID *int) (*Apartment, error)
	GetByOwnerID(ownerID int) ([]*Apartment, error)
	GetAll(filters map[string]interface{}, page, pageSize int) ([]*Apartment, int, error)
	GetAllWithUserContext(filters map[string]interface{}, page, pageSize int, userID *int) ([]*Apartment, int, error)
	GetByCoordinates(minLat, maxLat, minLng, maxLng float64) ([]*ApartmentCoordinates, error)
	GetByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*ApartmentCoordinates, error)
	GetFullApartmentsByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*Apartment, error)
	GetStatusStatistics() (map[string]int, error)
	GetCityStatistics() (map[string]int, error)
	GetDistrictStatistics() (map[string]int, error)
	Update(apartment *Apartment) error
	Delete(id int) error
	DeleteByAdmin(id int, force bool, adminID int) (hasActiveBookings bool, activeBookingsCount int, err error)

	AddPhotos(apartmentID int, filesData [][]byte) ([]string, error)
	AddPhotosParallel(apartmentID int, filesData [][]byte) ([]string, error)
	GetPhotosByApartmentID(apartmentID int) ([]*ApartmentPhoto, error)
	DeletePhoto(id int) error

	AddDocuments(apartmentID int, filesData [][]byte) ([]string, error)
	AddDocumentsWithType(apartmentID int, filesData [][]byte, documentType string) ([]string, error)
	GetDocumentsByApartmentID(apartmentID int) ([]*ApartmentDocument, error)
	GetDocumentsByApartmentIDAndType(apartmentID int, documentType string) ([]*ApartmentDocument, error)
	GetDocumentByID(id int) (*ApartmentDocument, error)
	DeleteDocument(id int) error

	AddLocation(location *ApartmentLocation) error
	UpdateLocation(location *ApartmentLocation) error
	GetLocationByApartmentID(apartmentID int) (*ApartmentLocation, error)

	GetAllConditions() ([]*ApartmentCondition, error)
	GetConditionByID(id int) (*ApartmentCondition, error)

	UpdateStatus(apartmentID int, status ApartmentStatus, comment string) error
	UpdateApartmentType(apartmentID int, apartmentTypeID int) error

	GetAllHouseRules() ([]*HouseRules, error)
	GetHouseRulesByID(id int) (*HouseRules, error)
	GetAllPopularAmenities() ([]*PopularAmenities, error)
	GetPopularAmenitiesByID(id int) (*PopularAmenities, error)

	AddHouseRulesToApartment(apartmentID int, houseRuleIDs []int) error
	AddAmenitiesToApartment(apartmentID int, amenityIDs []int) error
	GetHouseRulesByApartmentID(apartmentID int) ([]*HouseRules, error)
	GetAmenitiesByApartmentID(apartmentID int) ([]*PopularAmenities, error)

	GetBookedDates(apartmentID int, daysAhead int) ([]string, error)

	ConfirmApartmentAgreement(apartmentID, userID int, request *ConfirmApartmentAgreementRequest) (*Apartment, error)

	IncrementViewCount(apartmentID int) error
	IncrementBookingCount(apartmentID int) error
	AdminUpdateViewCount(apartmentID int, viewCount int) error
	AdminUpdateBookingCount(apartmentID int, bookingCount int) error
	AdminResetCounters(apartmentID int) error
}

type ApartmentAvailabilityService interface {
	RecalculateApartmentAvailability(apartmentID int) error
	RecalculateMultipleApartments(apartmentIDs []int) error
	RecalculateAllApartments() error
	CleanupExpiredCreatedBookings() error
}

type HouseRules struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PopularAmenities struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ConfirmApartmentAgreementRequest struct {
	IsAgreementAccepted bool `json:"is_agreement_accepted" validate:"required"`
}

type ApartmentResponse struct {
	ID                   int                  `json:"id"`
	OwnerID              int                  `json:"owner_id"`
	Owner                *PropertyOwner       `json:"owner,omitempty"`
	CityID               int                  `json:"city_id"`
	City                 *City                `json:"city,omitempty"`
	DistrictID           int                  `json:"district_id"`
	District             *District            `json:"district,omitempty"`
	MicrodistrictID      *int                 `json:"microdistrict_id,omitempty"`
	Microdistrict        *Microdistrict       `json:"microdistrict,omitempty"`
	ApartmentTypeID      *int                 `json:"apartment_type_id,omitempty"`
	ApartmentType        *ApartmentType       `json:"apartment_type,omitempty"`
	Street               string               `json:"street"`
	Building             string               `json:"building"`
	ApartmentNumber      int                  `json:"apartment_number"`
	ResidentialComplex   *string              `json:"residential_complex,omitempty"`
	RoomCount            int                  `json:"room_count"`
	TotalArea            float64              `json:"total_area"`
	KitchenArea          float64              `json:"kitchen_area"`
	Floor                int                  `json:"floor"`
	TotalFloors          int                  `json:"total_floors"`
	ConditionID          int                  `json:"condition_id"`
	Condition            *ApartmentCondition  `json:"condition,omitempty"`
	Price                int                  `json:"price"`
	DailyPrice           int                  `json:"daily_price"`
	ServiceFeePercentage int                  `json:"service_fee_percentage"`
	RentalTypeHourly     bool                 `json:"rental_type_hourly"`
	RentalTypeDaily      bool                 `json:"rental_type_daily"`
	IsFree               bool                 `json:"is_free"`
	IsFavorite           bool                 `json:"is_favorite"`
	Status               ApartmentStatus      `json:"status"`
	ModeratorComment     string               `json:"moderator_comment,omitempty"`
	Description          string               `json:"description"`
	ListingType          string               `json:"listing_type"`
	IsAgreementAccepted  bool                 `json:"is_agreement_accepted"`
	AgreementAcceptedAt  *time.Time           `json:"agreement_accepted_at,omitempty"`
	ContractID           *int                 `json:"contract_id,omitempty"`
	HouseRules           []*HouseRules        `json:"house_rules,omitempty"`
	Amenities            []*PopularAmenities  `json:"amenities,omitempty"`
	Photos               []*ApartmentPhoto    `json:"photos,omitempty"`
	Documents            []*ApartmentDocument `json:"documents,omitempty"`
	Location             *ApartmentLocation   `json:"location,omitempty"`
	ViewCount            int                  `json:"view_count"`
	BookingCount         int                  `json:"booking_count"`
	CreatedAt            string               `json:"created_at"`
	UpdatedAt            string               `json:"updated_at"`
}

type UpdateApartmentTypeIDRequest struct {
	ApartmentTypeID *int `json:"apartment_type_id" validate:"required"`
}

type AdminUpdateCountersRequest struct {
	ViewCount    *int `json:"view_count,omitempty" validate:"omitempty,min=0"`
	BookingCount *int `json:"booking_count,omitempty" validate:"omitempty,min=0"`
}
