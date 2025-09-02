package utils

import (
	"database/sql"

	"github.com/russo2642/renti_kz/internal/domain"
)

const ApartmentSelectFields = `
	a.id, a.owner_id, a.city_id, a.district_id, a.microdistrict_id,
	a.street, a.building, a.apartment_number, a.residential_complex, a.room_count,
	a.total_area, a.kitchen_area, a.floor, a.total_floors,
	a.condition_id, a.price, a.daily_price, a.rental_type_hourly, a.rental_type_daily,
	a.is_free, a.status, a.moderator_comment, a.description, a.listing_type,
	a.is_agreement_accepted, a.agreement_accepted_at, a.contract_id, a.apartment_type_id,
	a.view_count, a.booking_count, a.created_at, a.updated_at`

const ApartmentWithConditionSelectFields = ApartmentSelectFields + `,
	c.name as condition_name, c.description as condition_description,
	at.id as apartment_type_id_scan, at.name as apartment_type_name, at.description as apartment_type_description`

const ApartmentWithOwnerSelectFields = ApartmentSelectFields + `,
	po.id as owner_id_real, po.user_id as owner_user_id, po.created_at as owner_created_at, po.updated_at as owner_updated_at,
	u.id as user_id, u.phone as user_phone, u.first_name as user_first_name, u.last_name as user_last_name, 
	u.email as user_email, u.city_id as user_city_id, u.iin as user_iin, u.role_id as user_role_id, 
	u.created_at as user_created_at, u.updated_at as user_updated_at`

const ApartmentWithConditionAndOwnerSelectFields = ApartmentSelectFields + `,
	c.name as condition_name, c.description as condition_description,
	po.id as owner_id_real, po.user_id as owner_user_id, po.created_at as owner_created_at, po.updated_at as owner_updated_at,
	u.id as user_id, u.phone as user_phone, u.first_name as user_first_name, u.last_name as user_last_name, 
	u.email as user_email, u.city_id as user_city_id, u.iin as user_iin, u.role_id as user_role_id, 
	u.created_at as user_created_at, u.updated_at as user_updated_at,
	at.id as apartment_type_id_scan, at.name as apartment_type_name, at.description as apartment_type_description`

const ApartmentWithConditionOwnerAndFavoriteSelectFields = ApartmentSelectFields + `,
	c.name as condition_name, c.description as condition_description,
	po.id as owner_id_real, po.user_id as owner_user_id, po.created_at as owner_created_at, po.updated_at as owner_updated_at,
	u.id as user_id, u.phone as user_phone, u.first_name as user_first_name, u.last_name as user_last_name, 
	u.email as user_email, u.city_id as user_city_id, u.iin as user_iin, u.role_id as user_role_id, 
	u.created_at as user_created_at, u.updated_at as user_updated_at,
	at.id as apartment_type_id_scan, at.name as apartment_type_name, at.description as apartment_type_description,
	CASE WHEN f.id IS NOT NULL THEN true ELSE false END as is_favorite`

func ScanApartment(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Apartment, error) {
	apartment := &domain.Apartment{}
	var microdistrictID sql.NullInt64
	var moderatorComment sql.NullString
	var description sql.NullString
	var residentialComplex sql.NullString
	var agreementAcceptedAt sql.NullTime
	var contractID, apartmentTypeID sql.NullInt32

	err := scanner.Scan(
		&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &microdistrictID,
		&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &residentialComplex, &apartment.RoomCount,
		&apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor, &apartment.TotalFloors,
		&apartment.ConditionID, &apartment.Price, &apartment.DailyPrice, &apartment.RentalTypeHourly, &apartment.RentalTypeDaily,
		&apartment.IsFree, &apartment.Status, &moderatorComment, &description, &apartment.ListingType,
		&apartment.IsAgreementAccepted, &agreementAcceptedAt, &contractID, &apartmentTypeID,
		&apartment.ViewCount, &apartment.BookingCount, &apartment.CreatedAt, &apartment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if microdistrictID.Valid {
		microId := int(microdistrictID.Int64)
		apartment.MicrodistrictID = &microId
	}

	if moderatorComment.Valid {
		apartment.ModeratorComment = moderatorComment.String
	}

	if description.Valid {
		apartment.Description = description.String
	}

	if residentialComplex.Valid {
		apartment.ResidentialComplex = &residentialComplex.String
	}

	if agreementAcceptedAt.Valid {
		apartment.AgreementAcceptedAt = &agreementAcceptedAt.Time
	}

	if contractID.Valid {
		id := int(contractID.Int32)
		apartment.ContractID = &id
	}

	if apartmentTypeID.Valid {
		typeID := int(apartmentTypeID.Int32)
		apartment.ApartmentTypeID = &typeID
	}

	apartment.ServiceFeePercentage = domain.ServiceFeePercentage

	return apartment, nil
}

func ScanApartmentWithCondition(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Apartment, error) {
	apartment := &domain.Apartment{}
	var microdistrictID sql.NullInt64
	var moderatorComment sql.NullString
	var description sql.NullString
	var residentialComplex sql.NullString
	var condition domain.ApartmentCondition
	var apartmentTypeName, apartmentTypeDescription sql.NullString
	var agreementAcceptedAt sql.NullTime
	var contractID, apartmentTypeIDScan sql.NullInt32

	err := scanner.Scan(
		&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &microdistrictID,
		&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &residentialComplex, &apartment.RoomCount,
		&apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor, &apartment.TotalFloors,
		&apartment.ConditionID, &apartment.Price, &apartment.DailyPrice, &apartment.RentalTypeHourly, &apartment.RentalTypeDaily,
		&apartment.IsFree, &apartment.Status, &moderatorComment, &description, &apartment.ListingType,
		&apartment.IsAgreementAccepted, &agreementAcceptedAt, &contractID, &apartment.ApartmentTypeID,
		&apartment.ViewCount, &apartment.BookingCount, &apartment.CreatedAt, &apartment.UpdatedAt, &condition.Name, &condition.Description,
		&apartmentTypeIDScan, &apartmentTypeName, &apartmentTypeDescription,
	)

	if err != nil {
		return nil, err
	}

	if microdistrictID.Valid {
		microId := int(microdistrictID.Int64)
		apartment.MicrodistrictID = &microId
	}

	if moderatorComment.Valid {
		apartment.ModeratorComment = moderatorComment.String
	}

	if description.Valid {
		apartment.Description = description.String
	}

	if residentialComplex.Valid {
		apartment.ResidentialComplex = &residentialComplex.String
	}

	if agreementAcceptedAt.Valid {
		apartment.AgreementAcceptedAt = &agreementAcceptedAt.Time
	}

	if contractID.Valid {
		id := int(contractID.Int32)
		apartment.ContractID = &id
	}

	condition.ID = apartment.ConditionID
	apartment.Condition = &condition

	if apartmentTypeIDScan.Valid {
		typeID := int(apartmentTypeIDScan.Int32)
		apartment.ApartmentTypeID = &typeID
		if apartmentTypeName.Valid || apartmentTypeDescription.Valid {
			apartment.ApartmentType = &domain.ApartmentType{
				ID:          typeID,
				Name:        apartmentTypeName.String,
				Description: apartmentTypeDescription.String,
			}
		}
	}

	apartment.ServiceFeePercentage = domain.ServiceFeePercentage

	return apartment, nil
}

func ScanApartmentWithOwner(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Apartment, error) {
	apartment := &domain.Apartment{}
	var microdistrictID sql.NullInt64
	var moderatorComment sql.NullString
	var description sql.NullString
	var residentialComplex sql.NullString
	var owner domain.PropertyOwner
	var user domain.User
	var agreementAcceptedAt sql.NullTime
	var contractID sql.NullInt32

	err := scanner.Scan(
		&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &microdistrictID,
		&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &residentialComplex, &apartment.RoomCount,
		&apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor, &apartment.TotalFloors,
		&apartment.ConditionID, &apartment.Price, &apartment.DailyPrice, &apartment.RentalTypeHourly, &apartment.RentalTypeDaily,
		&apartment.IsFree, &apartment.Status, &moderatorComment, &description, &apartment.ListingType,
		&apartment.IsAgreementAccepted, &agreementAcceptedAt, &contractID, &apartment.ApartmentTypeID,
		&apartment.ViewCount, &apartment.BookingCount, &apartment.CreatedAt, &apartment.UpdatedAt, &owner.ID, &owner.UserID, &owner.CreatedAt, &owner.UpdatedAt,
		&user.ID, &user.Phone, &user.FirstName, &user.LastName,
		&user.Email, &user.CityID, &user.IIN, &user.RoleID,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if microdistrictID.Valid {
		microId := int(microdistrictID.Int64)
		apartment.MicrodistrictID = &microId
	}

	if moderatorComment.Valid {
		apartment.ModeratorComment = moderatorComment.String
	}

	if description.Valid {
		apartment.Description = description.String
	}

	if residentialComplex.Valid {
		apartment.ResidentialComplex = &residentialComplex.String
	}

	if agreementAcceptedAt.Valid {
		apartment.AgreementAcceptedAt = &agreementAcceptedAt.Time
	}

	if contractID.Valid {
		id := int(contractID.Int32)
		apartment.ContractID = &id
	}

	switch user.RoleID {
	case 1:
		user.Role = domain.RoleUser
	case 2:
		user.Role = domain.RoleOwner
	case 3:
		user.Role = domain.RoleModerator
	case 4:
		user.Role = domain.RoleAdmin
	case 5:
		user.Role = domain.RoleConcierge
	default:
		user.Role = domain.RoleUser
	}

	owner.User = &user
	apartment.Owner = &owner

	apartment.ServiceFeePercentage = domain.ServiceFeePercentage

	return apartment, nil
}

func ScanApartmentWithConditionAndOwner(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Apartment, error) {
	apartment := &domain.Apartment{}
	var microdistrictID sql.NullInt64
	var moderatorComment sql.NullString
	var description sql.NullString
	var residentialComplex sql.NullString
	var condition domain.ApartmentCondition
	var apartmentTypeName, apartmentTypeDescription sql.NullString
	var owner domain.PropertyOwner
	var user domain.User
	var agreementAcceptedAt sql.NullTime
	var contractID, apartmentTypeIDScan sql.NullInt32

	err := scanner.Scan(
		&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &microdistrictID,
		&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &residentialComplex, &apartment.RoomCount,
		&apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor, &apartment.TotalFloors,
		&apartment.ConditionID, &apartment.Price, &apartment.DailyPrice, &apartment.RentalTypeHourly, &apartment.RentalTypeDaily,
		&apartment.IsFree, &apartment.Status, &moderatorComment, &description, &apartment.ListingType,
		&apartment.IsAgreementAccepted, &agreementAcceptedAt, &contractID, &apartment.ApartmentTypeID,
		&apartment.ViewCount, &apartment.BookingCount, &apartment.CreatedAt, &apartment.UpdatedAt, &condition.Name, &condition.Description,
		&owner.ID, &owner.UserID, &owner.CreatedAt, &owner.UpdatedAt,
		&user.ID, &user.Phone, &user.FirstName, &user.LastName,
		&user.Email, &user.CityID, &user.IIN, &user.RoleID,
		&user.CreatedAt, &user.UpdatedAt,
		&apartmentTypeIDScan, &apartmentTypeName, &apartmentTypeDescription,
	)

	if err != nil {
		return nil, err
	}

	if microdistrictID.Valid {
		microId := int(microdistrictID.Int64)
		apartment.MicrodistrictID = &microId
	}

	if moderatorComment.Valid {
		apartment.ModeratorComment = moderatorComment.String
	}

	if description.Valid {
		apartment.Description = description.String
	}

	if agreementAcceptedAt.Valid {
		apartment.AgreementAcceptedAt = &agreementAcceptedAt.Time
	}

	if contractID.Valid {
		id := int(contractID.Int32)
		apartment.ContractID = &id
	}

	condition.ID = apartment.ConditionID
	apartment.Condition = &condition

	switch user.RoleID {
	case 1:
		user.Role = domain.RoleUser
	case 2:
		user.Role = domain.RoleOwner
	case 3:
		user.Role = domain.RoleModerator
	case 4:
		user.Role = domain.RoleAdmin
	case 5:
		user.Role = domain.RoleConcierge
	default:
		user.Role = domain.RoleUser
	}

	if residentialComplex.Valid {
		apartment.ResidentialComplex = &residentialComplex.String
	}

	owner.User = &user
	apartment.Owner = &owner

	apartment.ServiceFeePercentage = domain.ServiceFeePercentage

	return apartment, nil
}

func ScanApartmentWithConditionOwnerAndFavorite(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Apartment, error) {
	apartment := &domain.Apartment{}
	var microdistrictID sql.NullInt64
	var moderatorComment sql.NullString
	var description sql.NullString
	var residentialComplex sql.NullString
	var condition domain.ApartmentCondition
	var apartmentTypeName, apartmentTypeDescription sql.NullString
	var owner domain.PropertyOwner
	var user domain.User
	var agreementAcceptedAt sql.NullTime
	var contractID, apartmentTypeIDScan sql.NullInt32

	err := scanner.Scan(
		&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &microdistrictID,
		&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &residentialComplex, &apartment.RoomCount,
		&apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor, &apartment.TotalFloors,
		&apartment.ConditionID, &apartment.Price, &apartment.DailyPrice, &apartment.RentalTypeHourly, &apartment.RentalTypeDaily,
		&apartment.IsFree, &apartment.Status, &moderatorComment, &description, &apartment.ListingType,
		&apartment.IsAgreementAccepted, &agreementAcceptedAt, &contractID, &apartment.ApartmentTypeID,
		&apartment.ViewCount, &apartment.BookingCount, &apartment.CreatedAt, &apartment.UpdatedAt, &condition.Name, &condition.Description,
		&owner.ID, &owner.UserID, &owner.CreatedAt, &owner.UpdatedAt,
		&user.ID, &user.Phone, &user.FirstName, &user.LastName,
		&user.Email, &user.CityID, &user.IIN, &user.RoleID,
		&user.CreatedAt, &user.UpdatedAt,
		&apartmentTypeIDScan, &apartmentTypeName, &apartmentTypeDescription,
		&apartment.IsFavorite,
	)

	if err != nil {
		return nil, err
	}

	if microdistrictID.Valid {
		microId := int(microdistrictID.Int64)
		apartment.MicrodistrictID = &microId
	}

	if moderatorComment.Valid {
		apartment.ModeratorComment = moderatorComment.String
	}

	if description.Valid {
		apartment.Description = description.String
	}

	if agreementAcceptedAt.Valid {
		apartment.AgreementAcceptedAt = &agreementAcceptedAt.Time
	}

	if contractID.Valid {
		id := int(contractID.Int32)
		apartment.ContractID = &id
	}

	condition.ID = apartment.ConditionID
	apartment.Condition = &condition

	switch user.RoleID {
	case 1:
		user.Role = domain.RoleUser
	case 2:
		user.Role = domain.RoleOwner
	case 3:
		user.Role = domain.RoleModerator
	case 4:
		user.Role = domain.RoleAdmin
	case 5:
		user.Role = domain.RoleConcierge
	default:
		user.Role = domain.RoleUser
	}

	if residentialComplex.Valid {
		apartment.ResidentialComplex = &residentialComplex.String
	}

	if apartmentTypeIDScan.Valid {
		typeID := int(apartmentTypeIDScan.Int32)
		apartment.ApartmentTypeID = &typeID
		if apartmentTypeName.Valid || apartmentTypeDescription.Valid {
			apartment.ApartmentType = &domain.ApartmentType{
				ID:          typeID,
				Name:        apartmentTypeName.String,
				Description: apartmentTypeDescription.String,
			}
		}
	}

	owner.User = &user
	apartment.Owner = &owner

	apartment.ServiceFeePercentage = domain.ServiceFeePercentage

	return apartment, nil
}

func ScanApartmentsWithConditionAndOwner(rows *sql.Rows) ([]*domain.Apartment, error) {
	var apartments []*domain.Apartment

	for rows.Next() {
		apartment, err := ScanApartmentWithConditionAndOwner(rows)
		if err != nil {
			return nil, err
		}
		apartments = append(apartments, apartment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apartments, nil
}

func ScanApartmentsWithConditionOwnerAndFavorite(rows *sql.Rows) ([]*domain.Apartment, error) {
	var apartments []*domain.Apartment

	for rows.Next() {
		apartment, err := ScanApartmentWithConditionOwnerAndFavorite(rows)
		if err != nil {
			return nil, err
		}
		apartments = append(apartments, apartment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apartments, nil
}

func ScanApartments(rows *sql.Rows) ([]*domain.Apartment, error) {
	var apartments []*domain.Apartment

	for rows.Next() {
		apartment, err := ScanApartment(rows)
		if err != nil {
			return nil, err
		}
		apartments = append(apartments, apartment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apartments, nil
}

const ApartmentPhotoSelectFields = `
	id, apartment_id, url, created_at`

func ScanApartmentPhoto(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.ApartmentPhoto, error) {
	photo := &domain.ApartmentPhoto{}

	err := scanner.Scan(
		&photo.ID, &photo.ApartmentID, &photo.URL, &photo.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return photo, nil
}

func ScanApartmentPhotos(rows *sql.Rows) ([]*domain.ApartmentPhoto, error) {
	var photos []*domain.ApartmentPhoto

	for rows.Next() {
		photo, err := ScanApartmentPhoto(rows)
		if err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return photos, nil
}

const ApartmentLocationSelectFields = `
	id, apartment_id, latitude, longitude, created_at, updated_at`

func ScanApartmentLocation(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.ApartmentLocation, error) {
	location := &domain.ApartmentLocation{}

	err := scanner.Scan(
		&location.ID, &location.ApartmentID, &location.Latitude, &location.Longitude,
		&location.CreatedAt, &location.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return location, nil
}

const ApartmentConditionSelectFields = `
	id, name, description`

func ScanApartmentCondition(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.ApartmentCondition, error) {
	condition := &domain.ApartmentCondition{}

	err := scanner.Scan(
		&condition.ID, &condition.Name, &condition.Description,
	)

	if err != nil {
		return nil, err
	}

	return condition, nil
}

func ScanApartmentConditions(rows *sql.Rows) ([]*domain.ApartmentCondition, error) {
	var conditions []*domain.ApartmentCondition

	for rows.Next() {
		condition, err := ScanApartmentCondition(rows)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, condition)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conditions, nil
}

const HouseRulesSelectFields = `
	id, name, description`

func ScanHouseRules(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.HouseRules, error) {
	rules := &domain.HouseRules{}

	err := scanner.Scan(
		&rules.ID, &rules.Name, &rules.Description,
	)

	if err != nil {
		return nil, err
	}

	return rules, nil
}

func ScanMultipleHouseRules(rows *sql.Rows) ([]*domain.HouseRules, error) {
	var rules []*domain.HouseRules

	for rows.Next() {
		rule, err := ScanHouseRules(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

const PopularAmenitiesSelectFields = `
	id, name, description`

func ScanPopularAmenities(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.PopularAmenities, error) {
	amenities := &domain.PopularAmenities{}

	err := scanner.Scan(
		&amenities.ID, &amenities.Name, &amenities.Description,
	)

	if err != nil {
		return nil, err
	}

	return amenities, nil
}

func ScanMultiplePopularAmenities(rows *sql.Rows) ([]*domain.PopularAmenities, error) {
	var amenities []*domain.PopularAmenities

	for rows.Next() {
		amenity, err := ScanPopularAmenities(rows)
		if err != nil {
			return nil, err
		}
		amenities = append(amenities, amenity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return amenities, nil
}

const ApartmentDocumentSelectFields = `
	id, apartment_id, document_type, url, created_at`

func ScanApartmentDocument(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.ApartmentDocument, error) {
	document := &domain.ApartmentDocument{}

	err := scanner.Scan(
		&document.ID, &document.ApartmentID, &document.Type, &document.URL, &document.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func ScanApartmentDocuments(rows *sql.Rows) ([]*domain.ApartmentDocument, error) {
	var documents []*domain.ApartmentDocument

	for rows.Next() {
		document, err := ScanApartmentDocument(rows)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}
