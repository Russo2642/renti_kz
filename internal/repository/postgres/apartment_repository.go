package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type ApartmentRepository struct {
	db *sql.DB
}

func NewApartmentRepository(db *sql.DB) *ApartmentRepository {
	return &ApartmentRepository{
		db: db,
	}
}

func (r *ApartmentRepository) Create(apartment *domain.Apartment) error {
	query := `
		INSERT INTO apartments (
			owner_id, city_id, district_id, microdistrict_id, 
			street, building, apartment_number, residential_complex, room_count, 
			total_area, kitchen_area, floor, total_floors, 
			condition_id, price, daily_price, rental_type_hourly, rental_type_daily, 
			is_free, status, moderator_comment, description, listing_type,
			is_agreement_accepted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		apartment.OwnerID,
		apartment.CityID,
		apartment.DistrictID,
		apartment.MicrodistrictID,
		apartment.Street,
		apartment.Building,
		apartment.ApartmentNumber,
		apartment.ResidentialComplex,
		apartment.RoomCount,
		apartment.TotalArea,
		apartment.KitchenArea,
		apartment.Floor,
		apartment.TotalFloors,
		apartment.ConditionID,
		apartment.Price,
		apartment.DailyPrice,
		apartment.RentalTypeHourly,
		apartment.RentalTypeDaily,
		apartment.IsFree,
		apartment.Status,
		apartment.ModeratorComment,
		apartment.Description,
		apartment.ListingType,
		apartment.IsAgreementAccepted,
	).Scan(&apartment.ID, &apartment.CreatedAt, &apartment.UpdatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "apartment", "create")
	}

	if len(apartment.HouseRules) > 0 {
		houseRuleIDs := make([]int, 0, len(apartment.HouseRules))
		for _, rule := range apartment.HouseRules {
			houseRuleIDs = append(houseRuleIDs, rule.ID)
		}
		if err := r.AddHouseRulesToApartment(apartment.ID, houseRuleIDs); err != nil {
			return utils.HandleSQLError(err, "house rules", "add to apartment")
		}

		updatedRules, err := r.GetHouseRulesByApartmentID(apartment.ID)
		if err != nil {
			return utils.HandleSQLError(err, "house rules", "load after creation")
		}
		apartment.HouseRules = updatedRules
	}

	if len(apartment.Amenities) > 0 {
		amenityIDs := make([]int, 0, len(apartment.Amenities))
		for _, amenity := range apartment.Amenities {
			amenityIDs = append(amenityIDs, amenity.ID)
		}
		if err := r.AddAmenitiesToApartment(apartment.ID, amenityIDs); err != nil {
			return utils.HandleSQLError(err, "amenities", "add to apartment")
		}

		updatedAmenities, err := r.GetAmenitiesByApartmentID(apartment.ID)
		if err != nil {
			return utils.HandleSQLError(err, "amenities", "load after creation")
		}
		apartment.Amenities = updatedAmenities
	}

	return nil
}

func (r *ApartmentRepository) GetByID(id int) (*domain.Apartment, error) {
	query := `
		SELECT 
			a.id, a.owner_id, a.city_id, a.district_id, a.microdistrict_id,
			a.street, a.building, a.apartment_number, a.residential_complex, a.room_count,
			a.total_area, a.kitchen_area, a.floor, a.total_floors, a.condition_id,
			a.price, a.daily_price, a.rental_type_hourly, a.rental_type_daily,
			a.is_free, a.status, a.moderator_comment, a.description, a.listing_type,
			a.is_agreement_accepted, a.agreement_accepted_at, a.contract_id, a.apartment_type_id,
			a.view_count, a.booking_count, a.created_at, a.updated_at,
			po.id as owner_id, u.first_name as owner_first_name, u.last_name as owner_last_name,
			u.phone as owner_phone, u.email as owner_email, u.iin as owner_iin,
			c.name as city_name, d.name as district_name, m.name as microdistrict_name,
			ac.name as condition_name, ac.description as condition_description,
			at.id as apartment_type_id, at.name as apartment_type_name, at.description as apartment_type_description
		FROM apartments a
		LEFT JOIN property_owners po ON a.owner_id = po.id
		LEFT JOIN users u ON po.user_id = u.id
		LEFT JOIN cities c ON a.city_id = c.id
		LEFT JOIN districts d ON a.district_id = d.id
		LEFT JOIN microdistricts m ON a.microdistrict_id = m.id
		LEFT JOIN apartment_conditions ac ON a.condition_id = ac.id
		LEFT JOIN apartment_types at ON a.apartment_type_id = at.id
		WHERE a.id = $1`

	apartment := &domain.Apartment{}

	var ownerFirstName, ownerLastName, ownerPhone, ownerEmail, ownerIIN sql.NullString
	var cityName, districtName, microdistrictName sql.NullString
	var conditionName, conditionDescription sql.NullString
	var apartmentTypeName, apartmentTypeDescription sql.NullString
	var residentialComplex sql.NullString
	var microdistrictID, apartmentTypeID sql.NullInt32
	var agreementAcceptedAt sql.NullTime
	var contractID sql.NullInt32

	err := r.db.QueryRow(query, id).Scan(
		&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &microdistrictID,
		&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &residentialComplex, &apartment.RoomCount,
		&apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor, &apartment.TotalFloors, &apartment.ConditionID,
		&apartment.Price, &apartment.DailyPrice, &apartment.RentalTypeHourly, &apartment.RentalTypeDaily,
		&apartment.IsFree, &apartment.Status, &apartment.ModeratorComment, &apartment.Description, &apartment.ListingType,
		&apartment.IsAgreementAccepted, &agreementAcceptedAt, &contractID, &apartmentTypeID,
		&apartment.ViewCount, &apartment.BookingCount, &apartment.CreatedAt, &apartment.UpdatedAt,
		&apartment.OwnerID, &ownerFirstName, &ownerLastName,
		&ownerPhone, &ownerEmail, &ownerIIN,
		&cityName, &districtName, &microdistrictName,
		&conditionName, &conditionDescription,
		&apartmentTypeID, &apartmentTypeName, &apartmentTypeDescription,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("apartment with id %d not found", id)
		}
		return nil, utils.HandleSQLError(err, "apartment", "get by id")
	}

	if residentialComplex.Valid {
		apartment.ResidentialComplex = &residentialComplex.String
	}
	if microdistrictID.Valid {
		id := int(microdistrictID.Int32)
		apartment.MicrodistrictID = &id
	}
	if agreementAcceptedAt.Valid {
		apartment.AgreementAcceptedAt = &agreementAcceptedAt.Time
	}
	if contractID.Valid {
		id := int(contractID.Int32)
		apartment.ContractID = &id
	}

	apartment.ServiceFeePercentage = domain.ServiceFeePercentage

	if ownerFirstName.Valid {
		apartment.Owner = &domain.PropertyOwner{
			ID: apartment.OwnerID,
			User: &domain.User{
				FirstName: ownerFirstName.String,
				LastName:  ownerLastName.String,
				Phone:     ownerPhone.String,
				Email:     ownerEmail.String,
				IIN:       ownerIIN.String,
			},
		}
	}

	if cityName.Valid {
		apartment.City = &domain.City{
			ID:   apartment.CityID,
			Name: cityName.String,
		}
	}

	if districtName.Valid {
		apartment.District = &domain.District{
			ID:   apartment.DistrictID,
			Name: districtName.String,
		}
	}

	if microdistrictName.Valid && apartment.MicrodistrictID != nil {
		apartment.Microdistrict = &domain.Microdistrict{
			ID:   *apartment.MicrodistrictID,
			Name: microdistrictName.String,
		}
	}

	if conditionName.Valid {
		apartment.Condition = &domain.ApartmentCondition{
			ID:          apartment.ConditionID,
			Name:        conditionName.String,
			Description: conditionDescription.String,
		}
	}

	if apartmentTypeID.Valid {
		typeID := int(apartmentTypeID.Int32)
		apartment.ApartmentTypeID = &typeID
		apartment.ApartmentType = &domain.ApartmentType{
			ID:          typeID,
			Name:        apartmentTypeName.String,
			Description: apartmentTypeDescription.String,
		}
	}

	return apartment, nil
}

func (r *ApartmentRepository) GetByOwnerID(ownerID int) ([]*domain.Apartment, error) {
	query := `
		SELECT 
			a.id, a.owner_id, a.city_id, a.district_id, a.microdistrict_id, a.street, a.building, 
			a.apartment_number, a.residential_complex, a.room_count, a.total_area, a.kitchen_area, 
			a.floor, a.total_floors, a.condition_id, a.price, a.daily_price, a.rental_type_hourly, 
			a.rental_type_daily, a.is_free, a.status, a.description, a.listing_type,
			a.is_agreement_accepted, a.agreement_accepted_at, a.contract_id, a.apartment_type_id,
			a.created_at, a.updated_at
		FROM apartments a
		WHERE a.owner_id = $1
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment", "get by owner")
	}
	defer rows.Close()

	apartments := []*domain.Apartment{}
	for rows.Next() {
		apartment := &domain.Apartment{}
		var microdistrictID sql.NullInt32
		var residentialComplex sql.NullString
		var agreementAcceptedAt sql.NullTime
		var contractID, apartmentTypeID sql.NullInt32

		err := rows.Scan(
			&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &microdistrictID,
			&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &residentialComplex,
			&apartment.RoomCount, &apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor,
			&apartment.TotalFloors, &apartment.ConditionID, &apartment.Price, &apartment.DailyPrice,
			&apartment.RentalTypeHourly, &apartment.RentalTypeDaily, &apartment.IsFree,
			&apartment.Status, &apartment.Description, &apartment.ListingType,
			&apartment.IsAgreementAccepted, &agreementAcceptedAt, &contractID, &apartmentTypeID,
			&apartment.CreatedAt, &apartment.UpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment", "scan")
		}

		apartment.OwnerID = ownerID

		if microdistrictID.Valid {
			id := int(microdistrictID.Int32)
			apartment.MicrodistrictID = &id
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

		apartments = append(apartments, apartment)
	}

	if err = rows.Err(); err != nil {
		return nil, utils.HandleSQLError(err, "apartment", "rows iteration")
	}

	return apartments, nil
}

func (r *ApartmentRepository) buildFilterConditions(filters map[string]interface{}) ([]string, []interface{}, int) {
	var conditions []string
	var params []interface{}
	paramIndex := 1

	for key, value := range filters {
		switch key {
		case "owner_id":
			conditions = append(conditions, fmt.Sprintf("a.owner_id = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "city_id":
			conditions = append(conditions, fmt.Sprintf("a.city_id = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "district_id":
			conditions = append(conditions, fmt.Sprintf("a.district_id = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "microdistrict_id":
			conditions = append(conditions, fmt.Sprintf("a.microdistrict_id = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "room_count":
			conditions = append(conditions, fmt.Sprintf("a.room_count = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "apartment_type_id":
			conditions = append(conditions, fmt.Sprintf("a.apartment_type_id = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "min_area":
			conditions = append(conditions, fmt.Sprintf("a.total_area >= $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "max_area":
			conditions = append(conditions, fmt.Sprintf("a.total_area <= $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "min_price":
			conditions = append(conditions, fmt.Sprintf("a.price >= $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "max_price":
			conditions = append(conditions, fmt.Sprintf("a.price <= $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "status":
			conditions = append(conditions, fmt.Sprintf("a.status = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "is_free":
			conditions = append(conditions, fmt.Sprintf("a.is_free = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "rental_type_hourly":
			conditions = append(conditions, fmt.Sprintf("a.rental_type_hourly = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "rental_type_daily":
			conditions = append(conditions, fmt.Sprintf("a.rental_type_daily = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "listing_type":
			conditions = append(conditions, fmt.Sprintf("a.listing_type = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		}
	}

	return conditions, params, paramIndex
}

func (r *ApartmentRepository) buildSmartAvailabilityCondition() string {
	return `NOT EXISTS (
		SELECT 1 
		FROM bookings b 
		WHERE b.apartment_id = a.id 
		AND b.status IN ('pending', 'approved', 'active')
		AND (
			-- Проверяем пересечение с периодом: с ближайшего часа на 3 часа
			(b.start_date <= date_trunc('hour', NOW() AT TIME ZONE 'Asia/Almaty') + INTERVAL '1 hour' + INTERVAL '3 hours' - INTERVAL '1 minute'
			 AND (b.end_date + INTERVAL '1 minute' * b.cleaning_duration) > date_trunc('hour', NOW() AT TIME ZONE 'Asia/Almaty') + INTERVAL '1 hour') OR
			-- Проверяем если бронирование начинается внутри нашего окна
			(b.start_date >= date_trunc('hour', NOW() AT TIME ZONE 'Asia/Almaty') + INTERVAL '1 hour' 
			 AND b.start_date < date_trunc('hour', NOW() AT TIME ZONE 'Asia/Almaty') + INTERVAL '1 hour' + INTERVAL '3 hours') OR
			-- ВАЖНО: Проверяем что наше окно + время уборки не пересекается с началом существующего бронирования
			(date_trunc('hour', NOW() AT TIME ZONE 'Asia/Almaty') + INTERVAL '1 hour' + INTERVAL '3 hours' + INTERVAL '60 minutes' > b.start_date
			 AND date_trunc('hour', NOW() AT TIME ZONE 'Asia/Almaty') + INTERVAL '1 hour' < b.start_date)
		)
	)`
}

func (r *ApartmentRepository) addDefaultStatusFilter(filters map[string]interface{}, conditions []string, params []interface{}, paramIndex int) ([]string, []interface{}, int) {
	_, hasStatus := filters["status"]
	_, hasOwnerID := filters["owner_id"]
	_, includeAllStatuses := filters["include_all_statuses"]

	if !hasStatus && !hasOwnerID && !includeAllStatuses {
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", paramIndex))
		params = append(params, domain.AptStatusApproved)
		paramIndex++
	}

	return conditions, params, paramIndex
}

func (r *ApartmentRepository) buildWhereClause(conditions []string) string {
	if len(conditions) > 0 {
		return "WHERE " + strings.Join(conditions, " AND ")
	}
	return ""
}

func (r *ApartmentRepository) getApartmentCount(baseQuery, whereClause string, params []interface{}) (int, error) {
	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", baseQuery, whereClause)

	var total int
	err := r.db.QueryRow(countQuery, params...).Scan(&total)
	if err != nil {
		return 0, utils.HandleSQLError(err, "apartments count", "query")
	}

	return total, nil
}

func (r *ApartmentRepository) loadApartmentRelatedData(apartment *domain.Apartment) {
	if houseRules, err := r.GetHouseRulesByApartmentID(apartment.ID); err == nil {
		apartment.HouseRules = houseRules
	}

	if amenities, err := r.GetAmenitiesByApartmentID(apartment.ID); err == nil {
		apartment.Amenities = amenities
	}

	if photos, err := r.GetPhotosByApartmentID(apartment.ID); err == nil {
		apartment.Photos = photos
	}

	if location, err := r.GetLocationByApartmentID(apartment.ID); err == nil {
		apartment.Location = location
	}

}

func (r *ApartmentRepository) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.Apartment, int, error) {
	baseQuery := `
		FROM apartments a
		LEFT JOIN apartment_conditions c ON a.condition_id = c.id
		LEFT JOIN property_owners po ON a.owner_id = po.id
		LEFT JOIN users u ON po.user_id = u.id
		LEFT JOIN apartment_types at ON a.apartment_type_id = at.id
	`

	conditions, params, paramIndex := r.buildFilterConditions(filters)
	conditions, params, paramIndex = r.addDefaultStatusFilter(filters, conditions, params, paramIndex)
	whereClause := r.buildWhereClause(conditions)

	total, err := r.getApartmentCount(baseQuery, whereClause, params)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT `+utils.ApartmentWithConditionAndOwnerSelectFields+`
		%s %s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, paramIndex, paramIndex+1)

	params = append(params, pageSize, offset)

	rows, err := r.db.Query(dataQuery, params...)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "apartments", "query")
	}
	defer utils.CloseRows(rows)

	apartments, err := utils.ScanApartmentsWithConditionAndOwner(rows)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "apartments", "scan")
	}

	r.LoadApartmentsRelatedDataBatch(apartments)

	return apartments, total, nil
}

func (r *ApartmentRepository) Update(apartment *domain.Apartment) error {
	query := `
		UPDATE apartments SET 
			city_id = $2, district_id = $3, microdistrict_id = $4, apartment_type_id = $5,
			street = $6, building = $7, apartment_number = $8, residential_complex = $9,
			room_count = $10, total_area = $11, kitchen_area = $12, 
			floor = $13, total_floors = $14, condition_id = $15, 
			price = $16, daily_price = $17, rental_type_hourly = $18, rental_type_daily = $19,
			is_free = $20, status = $21, moderator_comment = $22, description = $23, listing_type = $24,
			is_agreement_accepted = $25, agreement_accepted_at = $26, contract_id = $27,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRow(
		query,
		apartment.ID,
		apartment.CityID,
		apartment.DistrictID,
		apartment.MicrodistrictID,
		apartment.ApartmentTypeID,
		apartment.Street,
		apartment.Building,
		apartment.ApartmentNumber,
		apartment.ResidentialComplex,
		apartment.RoomCount,
		apartment.TotalArea,
		apartment.KitchenArea,
		apartment.Floor,
		apartment.TotalFloors,
		apartment.ConditionID,
		apartment.Price,
		apartment.DailyPrice,
		apartment.RentalTypeHourly,
		apartment.RentalTypeDaily,
		apartment.IsFree,
		apartment.Status,
		apartment.ModeratorComment,
		apartment.Description,
		apartment.ListingType,
		apartment.IsAgreementAccepted,
		apartment.AgreementAcceptedAt,
		apartment.ContractID,
	).Scan(&apartment.UpdatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "apartment", "update")
	}

	if apartment.HouseRules != nil {
		houseRuleIDs := make([]int, 0, len(apartment.HouseRules))
		for _, rule := range apartment.HouseRules {
			houseRuleIDs = append(houseRuleIDs, rule.ID)
		}
		if err := r.AddHouseRulesToApartment(apartment.ID, houseRuleIDs); err != nil {
			return fmt.Errorf("failed to update house rules: %w", err)
		}

		updatedRules, err := r.GetHouseRulesByApartmentID(apartment.ID)
		if err != nil {
			return fmt.Errorf("failed to load house rules after update: %w", err)
		}
		apartment.HouseRules = updatedRules
	}

	if apartment.Amenities != nil {
		amenityIDs := make([]int, 0, len(apartment.Amenities))
		for _, amenity := range apartment.Amenities {
			amenityIDs = append(amenityIDs, amenity.ID)
		}
		if err := r.AddAmenitiesToApartment(apartment.ID, amenityIDs); err != nil {
			return fmt.Errorf("failed to update amenities: %w", err)
		}

		updatedAmenities, err := r.GetAmenitiesByApartmentID(apartment.ID)
		if err != nil {
			return fmt.Errorf("failed to load amenities after update: %w", err)
		}
		apartment.Amenities = updatedAmenities
	}

	return nil
}

func (r *ApartmentRepository) Delete(id int) error {
	query := "DELETE FROM apartments WHERE id = $1"

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete apartment: %w", err)
	}

	return nil
}

func (r *ApartmentRepository) AddPhoto(photo *domain.ApartmentPhoto) error {
	query := `
		INSERT INTO apartment_photos (apartment_id, url, "order")
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, photo.ApartmentID, photo.URL, photo.Order).Scan(
		&photo.ID, &photo.CreatedAt, &photo.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "apartment photo", "create")
	}

	return nil
}

func (r *ApartmentRepository) GetPhotosByApartmentID(apartmentID int) ([]*domain.ApartmentPhoto, error) {
	query := `
		SELECT id, apartment_id, url, "order", created_at, updated_at
		FROM apartment_photos
		WHERE apartment_id = $1
		ORDER BY "order" ASC
	`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment photos", "query")
	}
	defer utils.CloseRows(rows)

	var photos []*domain.ApartmentPhoto
	for rows.Next() {
		var photo domain.ApartmentPhoto

		err := rows.Scan(
			&photo.ID, &photo.ApartmentID, &photo.URL, &photo.Order,
			&photo.CreatedAt, &photo.UpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment photo", "scan")
		}

		photos = append(photos, &photo)
	}

	if err = utils.CheckRowsError(rows, "apartment photos iteration"); err != nil {
		return nil, err
	}

	return photos, nil
}

func (r *ApartmentRepository) GetPhotoByID(id int) (*domain.ApartmentPhoto, error) {
	query := `
		SELECT id, apartment_id, url, "order", created_at, updated_at
		FROM apartment_photos
		WHERE id = $1
	`

	var photo domain.ApartmentPhoto
	err := r.db.QueryRow(query, id).Scan(
		&photo.ID, &photo.ApartmentID, &photo.URL, &photo.Order,
		&photo.CreatedAt, &photo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLErrorWithID(err, "apartment photo", "get", id)
	}

	return &photo, nil
}

func (r *ApartmentRepository) DeletePhoto(id int) error {
	query := "DELETE FROM apartment_photos WHERE id = $1"

	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "apartment photo", "delete", id)
	}

	return nil
}

func (r *ApartmentRepository) AddLocation(location *domain.ApartmentLocation) error {
	query := `
		INSERT INTO apartment_locations (apartment_id, latitude, longitude)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, location.ApartmentID, location.Latitude, location.Longitude).Scan(
		&location.ID, &location.CreatedAt, &location.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "apartment location", "create")
	}

	return nil
}

func (r *ApartmentRepository) UpdateLocation(location *domain.ApartmentLocation) error {
	query := `
		UPDATE apartment_locations
		SET latitude = $1, longitude = $2, updated_at = CURRENT_TIMESTAMP
		WHERE apartment_id = $3
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, location.Latitude, location.Longitude, location.ApartmentID).Scan(
		&location.ID, &location.CreatedAt, &location.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "apartment location", "update")
	}

	return nil
}

func (r *ApartmentRepository) GetLocationByApartmentID(apartmentID int) (*domain.ApartmentLocation, error) {
	query := `
		SELECT id, apartment_id, latitude, longitude, created_at, updated_at
		FROM apartment_locations
		WHERE apartment_id = $1
	`

	var location domain.ApartmentLocation
	err := r.db.QueryRow(query, apartmentID).Scan(
		&location.ID, &location.ApartmentID, &location.Latitude, &location.Longitude,
		&location.CreatedAt, &location.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "apartment location", "get")
	}

	return &location, nil
}

func (r *ApartmentRepository) GetAllConditions() ([]*domain.ApartmentCondition, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM apartment_conditions
		ORDER BY id ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment conditions", "query")
	}
	defer utils.CloseRows(rows)

	var conditions []*domain.ApartmentCondition
	for rows.Next() {
		var condition domain.ApartmentCondition

		err := rows.Scan(
			&condition.ID, &condition.Name, &condition.Description,
			&condition.CreatedAt, &condition.UpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment condition", "scan")
		}

		conditions = append(conditions, &condition)
	}

	if err = utils.CheckRowsError(rows, "apartment conditions iteration"); err != nil {
		return nil, err
	}

	return conditions, nil
}

func (r *ApartmentRepository) GetConditionByID(id int) (*domain.ApartmentCondition, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM apartment_conditions
		WHERE id = $1
	`

	var condition domain.ApartmentCondition
	err := r.db.QueryRow(query, id).Scan(
		&condition.ID, &condition.Name, &condition.Description,
		&condition.CreatedAt, &condition.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLErrorWithID(err, "apartment condition", "get", id)
	}

	return &condition, nil
}

func (r *ApartmentRepository) CreateCondition(condition *domain.ApartmentCondition) error {
	query := `
		INSERT INTO apartment_conditions (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, condition.Name, condition.Description).Scan(
		&condition.ID, &condition.CreatedAt, &condition.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "apartment condition", "create")
	}

	return nil
}

func (r *ApartmentRepository) UpdateCondition(condition *domain.ApartmentCondition) error {
	query := `
		UPDATE apartment_conditions
		SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(query, condition.Name, condition.Description, condition.ID).Scan(
		&condition.CreatedAt, &condition.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLErrorWithID(err, "apartment condition", "update", condition.ID)
	}

	return nil
}

func (r *ApartmentRepository) GetHouseRulesByID(id int) (*domain.HouseRules, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM house_rules
		WHERE id = $1
	`

	var rules domain.HouseRules
	err := r.db.QueryRow(query, id).Scan(
		&rules.ID,
		&rules.Name,
		&rules.Description,
		&rules.CreatedAt,
		&rules.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get house rules: %w", err)
	}

	return &rules, nil
}

func (r *ApartmentRepository) GetPopularAmenitiesByID(id int) (*domain.PopularAmenities, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM popular_amenities
		WHERE id = $1
	`

	var amenities domain.PopularAmenities
	err := r.db.QueryRow(query, id).Scan(
		&amenities.ID,
		&amenities.Name,
		&amenities.Description,
		&amenities.CreatedAt,
		&amenities.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get popular amenities: %w", err)
	}

	return &amenities, nil
}

func (r *ApartmentRepository) GetAllHouseRules() ([]*domain.HouseRules, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM house_rules
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get house rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.HouseRules
	for rows.Next() {
		var rule domain.HouseRules
		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan house rule: %w", err)
		}

		rules = append(rules, &rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating house rules: %w", err)
	}

	return rules, nil
}

func (r *ApartmentRepository) GetAllPopularAmenities() ([]*domain.PopularAmenities, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM popular_amenities
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular amenities: %w", err)
	}
	defer rows.Close()

	var amenities []*domain.PopularAmenities
	for rows.Next() {
		var amenity domain.PopularAmenities
		err := rows.Scan(
			&amenity.ID,
			&amenity.Name,
			&amenity.Description,
			&amenity.CreatedAt,
			&amenity.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan popular amenity: %w", err)
		}

		amenities = append(amenities, &amenity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating popular amenities: %w", err)
	}

	return amenities, nil
}

func (r *ApartmentRepository) AddHouseRulesToApartment(apartmentID int, houseRuleIDs []int) error {
	return utils.ExecuteInTransaction(r.db, func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM apartment_house_rules WHERE apartment_id = $1", apartmentID)
		if err != nil {
			return utils.HandleSQLError(err, "existing house rules", "delete")
		}

		stmt, err := tx.Prepare("INSERT INTO apartment_house_rules (apartment_id, house_rule_id) VALUES ($1, $2)")
		if err != nil {
			return utils.HandleSQLError(err, "house rules statement", "prepare")
		}
		defer stmt.Close()

		for _, ruleID := range houseRuleIDs {
			_, err = stmt.Exec(apartmentID, ruleID)
			if err != nil {
				return utils.HandleSQLError(err, "house rule", "add")
			}
		}

		return nil
	})
}

func (r *ApartmentRepository) AddAmenitiesToApartment(apartmentID int, amenityIDs []int) error {
	return utils.ExecuteInTransaction(r.db, func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM apartment_amenities WHERE apartment_id = $1", apartmentID)
		if err != nil {
			return utils.HandleSQLError(err, "existing amenities", "delete")
		}

		stmt, err := tx.Prepare("INSERT INTO apartment_amenities (apartment_id, amenity_id) VALUES ($1, $2)")
		if err != nil {
			return utils.HandleSQLError(err, "amenities statement", "prepare")
		}
		defer stmt.Close()

		for _, amenityID := range amenityIDs {
			_, err = stmt.Exec(apartmentID, amenityID)
			if err != nil {
				return utils.HandleSQLError(err, "amenity", "add")
			}
		}

		return nil
	})
}

func (r *ApartmentRepository) GetHouseRulesByApartmentID(apartmentID int) ([]*domain.HouseRules, error) {
	query := `
		SELECT hr.id, hr.name, hr.description, hr.created_at, hr.updated_at
		FROM house_rules hr
		JOIN apartment_house_rules ahr ON hr.id = ahr.house_rule_id
		WHERE ahr.apartment_id = $1
		ORDER BY hr.id
	`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "house rules for apartment", "query")
	}
	defer utils.CloseRows(rows)

	var rules []*domain.HouseRules
	for rows.Next() {
		var rule domain.HouseRules
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description,
			&rule.CreatedAt, &rule.UpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "house rule", "scan")
		}

		rules = append(rules, &rule)
	}

	if err = utils.CheckRowsError(rows, "house rules iteration"); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *ApartmentRepository) GetAmenitiesByApartmentID(apartmentID int) ([]*domain.PopularAmenities, error) {
	query := `
		SELECT pa.id, pa.name, pa.description, pa.created_at, pa.updated_at
		FROM popular_amenities pa
		JOIN apartment_amenities aa ON pa.id = aa.amenity_id
		WHERE aa.apartment_id = $1
		ORDER BY pa.id
	`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "amenities for apartment", "query")
	}
	defer utils.CloseRows(rows)

	var amenities []*domain.PopularAmenities
	for rows.Next() {
		var amenity domain.PopularAmenities
		err := rows.Scan(
			&amenity.ID, &amenity.Name, &amenity.Description,
			&amenity.CreatedAt, &amenity.UpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "amenity", "scan")
		}

		amenities = append(amenities, &amenity)
	}

	if err = utils.CheckRowsError(rows, "amenities iteration"); err != nil {
		return nil, err
	}

	return amenities, nil
}

func (r *ApartmentRepository) AddDocument(document *domain.ApartmentDocument) error {
	query := `
		INSERT INTO apartment_documents (apartment_id, url, type)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, document.ApartmentID, document.URL, document.Type).Scan(
		&document.ID, &document.CreatedAt, &document.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "apartment document", "create")
	}

	return nil
}

func (r *ApartmentRepository) GetDocumentsByApartmentID(apartmentID int) ([]*domain.ApartmentDocument, error) {
	query := `
		SELECT id, apartment_id, url, type, created_at, updated_at
		FROM apartment_documents
		WHERE apartment_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment documents", "query")
	}
	defer utils.CloseRows(rows)

	var documents []*domain.ApartmentDocument
	for rows.Next() {
		var document domain.ApartmentDocument

		err := rows.Scan(
			&document.ID, &document.ApartmentID, &document.URL, &document.Type,
			&document.CreatedAt, &document.UpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment document", "scan")
		}

		documents = append(documents, &document)
	}

	if err = utils.CheckRowsError(rows, "apartment documents iteration"); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *ApartmentRepository) GetDocumentByID(id int) (*domain.ApartmentDocument, error) {
	query := `
		SELECT id, apartment_id, url, type, created_at, updated_at
		FROM apartment_documents
		WHERE id = $1
	`

	var document domain.ApartmentDocument
	err := r.db.QueryRow(query, id).Scan(
		&document.ID, &document.ApartmentID, &document.URL, &document.Type,
		&document.CreatedAt, &document.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLErrorWithID(err, "apartment document", "get", id)
	}

	return &document, nil
}

func (r *ApartmentRepository) DeleteDocument(id int) error {
	query := "DELETE FROM apartment_documents WHERE id = $1"

	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "apartment document", "delete", id)
	}

	return nil
}

func (r *ApartmentRepository) GetDocumentsByApartmentIDAndType(apartmentID int, documentType string) ([]*domain.ApartmentDocument, error) {
	query := `
		SELECT id, apartment_id, url, type, created_at, updated_at
		FROM apartment_documents
		WHERE apartment_id = $1 AND type = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, apartmentID, documentType)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment documents by type", "query")
	}
	defer utils.CloseRows(rows)

	var documents []*domain.ApartmentDocument
	for rows.Next() {
		var document domain.ApartmentDocument

		err := rows.Scan(
			&document.ID, &document.ApartmentID, &document.URL, &document.Type,
			&document.CreatedAt, &document.UpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment document", "scan")
		}

		documents = append(documents, &document)
	}

	if err = utils.CheckRowsError(rows, "apartment documents iteration"); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *ApartmentRepository) UpdateIsFree(apartmentID int, isFree bool) error {
	query := `UPDATE apartments SET is_free = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(query, apartmentID, isFree)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "apartment is_free status", "update", apartmentID)
	}
	return nil
}

func (r *ApartmentRepository) UpdateMultipleIsFree(apartmentStatusMap map[int]bool) error {
	if len(apartmentStatusMap) == 0 {
		return nil
	}

	var trueIDs, falseIDs []int

	for apartmentID, isFree := range apartmentStatusMap {
		if isFree {
			trueIDs = append(trueIDs, apartmentID)
		} else {
			falseIDs = append(falseIDs, apartmentID)
		}
	}

	if len(trueIDs) > 0 {
		query := `UPDATE apartments SET is_free = true, updated_at = CURRENT_TIMESTAMP WHERE id = ANY($1)`
		trueIDsArray := "{" + fmt.Sprintf("%d", trueIDs[0])
		for i := 1; i < len(trueIDs); i++ {
			trueIDsArray += fmt.Sprintf(",%d", trueIDs[i])
		}
		trueIDsArray += "}"

		_, err := r.db.Exec(query, trueIDsArray)
		if err != nil {
			return utils.HandleSQLError(err, "apartments batch is_free=true", "update")
		}
	}

	if len(falseIDs) > 0 {
		query := `UPDATE apartments SET is_free = false, updated_at = CURRENT_TIMESTAMP WHERE id = ANY($1)`
		falseIDsArray := "{" + fmt.Sprintf("%d", falseIDs[0])
		for i := 1; i < len(falseIDs); i++ {
			falseIDsArray += fmt.Sprintf(",%d", falseIDs[i])
		}
		falseIDsArray += "}"

		_, err := r.db.Exec(query, falseIDsArray)
		if err != nil {
			return utils.HandleSQLError(err, "apartments batch is_free=false", "update")
		}
	}

	return nil
}

func (r *ApartmentRepository) GetByIDWithUserContext(id int, userID *int) (*domain.Apartment, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT ` + utils.ApartmentWithConditionOwnerAndFavoriteSelectFields + `
			FROM apartments a
			LEFT JOIN apartment_conditions c ON a.condition_id = c.id
			LEFT JOIN property_owners po ON a.owner_id = po.id
			LEFT JOIN users u ON po.user_id = u.id
			LEFT JOIN apartment_types at ON a.apartment_type_id = at.id
			LEFT JOIN favorites f ON a.id = f.apartment_id AND f.user_id = $2
			WHERE a.id = $1`
		args = []interface{}{id, *userID}
	} else {
		query = `
			SELECT ` + utils.ApartmentWithConditionAndOwnerSelectFields + `, false as is_favorite
			FROM apartments a
			LEFT JOIN apartment_conditions c ON a.condition_id = c.id
			LEFT JOIN property_owners po ON a.owner_id = po.id
			LEFT JOIN users u ON po.user_id = u.id
			LEFT JOIN apartment_types at ON a.apartment_type_id = at.id
			WHERE a.id = $1`
		args = []interface{}{id}
	}

	var apartment *domain.Apartment
	var err error

	if userID != nil {
		apartment, err = utils.ScanApartmentWithConditionOwnerAndFavorite(r.db.QueryRow(query, args...))
	} else {
		apartment, err = utils.ScanApartmentWithConditionOwnerAndFavorite(r.db.QueryRow(query, args...))
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLErrorWithID(err, "apartment", "get", id)
	}

	houseRules, err := r.GetHouseRulesByApartmentID(apartment.ID)
	if err == nil {
		apartment.HouseRules = houseRules
	}

	amenities, err := r.GetAmenitiesByApartmentID(apartment.ID)
	if err == nil {
		apartment.Amenities = amenities
	}

	photos, err := r.GetPhotosByApartmentID(apartment.ID)
	if err == nil {
		apartment.Photos = photos
	}

	documents, err := r.GetDocumentsByApartmentID(apartment.ID)
	if err == nil {
		apartment.Documents = documents
	}

	location, err := r.GetLocationByApartmentID(apartment.ID)
	if err == nil {
		apartment.Location = location
	}

	return apartment, nil
}

func (r *ApartmentRepository) GetAllWithUserContext(filters map[string]interface{}, page, pageSize int, userID *int) ([]*domain.Apartment, int, error) {
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	var conditions []string

	if cityID, ok := filters["city_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.city_id = $%d", argIndex))
		args = append(args, cityID)
		argIndex++
	}

	if districtID, ok := filters["district_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.district_id = $%d", argIndex))
		args = append(args, districtID)
		argIndex++
	}

	if microdistrictID, ok := filters["microdistrict_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.microdistrict_id = $%d", argIndex))
		args = append(args, microdistrictID)
		argIndex++
	}

	if roomCount, ok := filters["room_count"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.room_count = $%d", argIndex))
		args = append(args, roomCount)
		argIndex++
	}

	if ownerID, ok := filters["owner_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.owner_id = $%d", argIndex))
		args = append(args, ownerID)
		argIndex++
	}

	if minArea, ok := filters["min_area"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.total_area >= $%d", argIndex))
		args = append(args, minArea)
		argIndex++
	}

	if maxArea, ok := filters["max_area"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.total_area <= $%d", argIndex))
		args = append(args, maxArea)
		argIndex++
	}

	if minPrice, ok := filters["min_price"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.price >= $%d", argIndex))
		args = append(args, minPrice)
		argIndex++
	}

	if maxPrice, ok := filters["max_price"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.price <= $%d", argIndex))
		args = append(args, maxPrice)
		argIndex++
	}

	if isFree, ok := filters["is_free"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.is_free = $%d", argIndex))
		args = append(args, isFree)
		argIndex++
	}

	if status, ok := filters["status"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	} else {
		conditions = append(conditions, "a.status = 'approved'")
	}

	if rentalTypeHourly, ok := filters["rental_type_hourly"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.rental_type_hourly = $%d", argIndex))
		args = append(args, rentalTypeHourly)
		argIndex++
	}

	if rentalTypeDaily, ok := filters["rental_type_daily"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.rental_type_daily = $%d", argIndex))
		args = append(args, rentalTypeDaily)
		argIndex++
	}

	if listingType, ok := filters["listing_type"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.listing_type = $%d", argIndex))
		args = append(args, listingType)
		argIndex++
	}

	if apartmentTypeID, ok := filters["apartment_type_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("a.apartment_type_id = $%d", argIndex))
		args = append(args, apartmentTypeID)
		argIndex++
	}

	whereClause = "WHERE " + strings.Join(conditions, " AND ")

	offset := (page - 1) * pageSize

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var countQuery string
	if userID != nil {
		countQuery = `
			SELECT COUNT(*) 
			FROM apartments a
			LEFT JOIN favorites f ON a.id = f.apartment_id AND f.user_id = $` + fmt.Sprintf("%d", argIndex) + `
			` + whereClause
		countArgs = append(countArgs, *userID)
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM apartments a
			` + whereClause
	}

	var totalCount int
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "apartment", "count")
	}

	var query string
	if userID != nil {
		query = `
			SELECT ` + utils.ApartmentWithConditionOwnerAndFavoriteSelectFields + `
			FROM apartments a
			LEFT JOIN apartment_conditions c ON a.condition_id = c.id
			LEFT JOIN property_owners po ON a.owner_id = po.id
			LEFT JOIN users u ON po.user_id = u.id
			LEFT JOIN apartment_types at ON a.apartment_type_id = at.id
			LEFT JOIN favorites f ON a.id = f.apartment_id AND f.user_id = $` + fmt.Sprintf("%d", argIndex) + `
			` + whereClause + `
			ORDER BY a.created_at DESC
			LIMIT $` + fmt.Sprintf("%d", argIndex+1) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+2)
		args = append(args, *userID, pageSize, offset)
	} else {
		query = `
			SELECT ` + utils.ApartmentWithConditionAndOwnerSelectFields + `, false as is_favorite
			FROM apartments a
			LEFT JOIN apartment_conditions c ON a.condition_id = c.id
			LEFT JOIN property_owners po ON a.owner_id = po.id
			LEFT JOIN users u ON po.user_id = u.id
			LEFT JOIN apartment_types at ON a.apartment_type_id = at.id
			` + whereClause + `
			ORDER BY a.created_at DESC
			LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
		args = append(args, pageSize, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "apartment", "get all")
	}
	defer rows.Close()

	apartments, err := utils.ScanApartmentsWithConditionOwnerAndFavorite(rows)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "apartment", "scan")
	}

	r.LoadApartmentsRelatedDataBatch(apartments)

	return apartments, totalCount, nil
}

func (r *ApartmentRepository) GetByCoordinates(minLat, maxLat, minLng, maxLng float64) ([]*domain.ApartmentCoordinates, error) {
	query := `
		SELECT 
			a.id,
			al.latitude,
			al.longitude
		FROM apartments a
		INNER JOIN apartment_locations al ON a.id = al.apartment_id
		WHERE al.latitude BETWEEN $1 AND $2 
		  AND al.longitude BETWEEN $3 AND $4
		  AND a.status = 'approved'
		ORDER BY a.id`

	rows, err := r.db.Query(query, minLat, maxLat, minLng, maxLng)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments by coordinates: %w", err)
	}
	defer rows.Close()

	var apartments []*domain.ApartmentCoordinates
	for rows.Next() {
		apartment := &domain.ApartmentCoordinates{}
		if err := rows.Scan(&apartment.ID, &apartment.Latitude, &apartment.Longitude); err != nil {
			return nil, fmt.Errorf("failed to scan apartment coordinates: %w", err)
		}
		apartments = append(apartments, apartment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return apartments, nil
}

func (r *ApartmentRepository) GetByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*domain.ApartmentCoordinates, error) {
	filterConditions, filterParams, _ := r.buildFilterConditions(filters)

	query := `
		SELECT 
			a.id,
			al.latitude,
			al.longitude
		FROM apartments a
		INNER JOIN apartment_locations al ON a.id = al.apartment_id
		WHERE al.latitude BETWEEN $1 AND $2 
		  AND al.longitude BETWEEN $3 AND $4`

	params := []interface{}{minLat, maxLat, minLng, maxLng}

	if len(filterConditions) > 0 {
		adjustedConditions := make([]string, len(filterConditions))
		for i, condition := range filterConditions {
			adjustedCondition := condition
			for j := len(filterParams); j >= 1; j-- {
				oldParam := fmt.Sprintf("$%d", j)
				newParam := fmt.Sprintf("$%d", j+4)
				adjustedCondition = strings.Replace(adjustedCondition, oldParam, newParam, 1)
			}
			adjustedConditions[i] = adjustedCondition
		}

		query += " AND " + strings.Join(adjustedConditions, " AND ")
		params = append(params, filterParams...)
	}

	query += " ORDER BY a.id"

	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments by coordinates with filters: %w", err)
	}
	defer rows.Close()

	var apartments []*domain.ApartmentCoordinates
	for rows.Next() {
		apartment := &domain.ApartmentCoordinates{}
		if err := rows.Scan(&apartment.ID, &apartment.Latitude, &apartment.Longitude); err != nil {
			return nil, fmt.Errorf("failed to scan apartment coordinates: %w", err)
		}
		apartments = append(apartments, apartment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return apartments, nil
}

func (r *ApartmentRepository) GetFullApartmentsByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*domain.Apartment, error) {
	baseQuery := `
		FROM apartments a
		LEFT JOIN apartment_conditions c ON a.condition_id = c.id
		LEFT JOIN property_owners po ON a.owner_id = po.id
		LEFT JOIN users u ON po.user_id = u.id
		LEFT JOIN apartment_types at ON a.apartment_type_id = at.id
		INNER JOIN apartment_locations al ON a.id = al.apartment_id
	`

	coordinateConditions := []string{
		"al.latitude BETWEEN $1 AND $2",
		"al.longitude BETWEEN $3 AND $4",
	}
	coordinateParams := []interface{}{minLat, maxLat, minLng, maxLng}

	filterConditions, filterParams, _ := r.buildFilterConditions(filters)

	allConditions := append(coordinateConditions, filterConditions...)
	allParams := append(coordinateParams, filterParams...)

	if len(filterConditions) > 0 {
		adjustedFilterConditions := make([]string, len(filterConditions))
		for i, condition := range filterConditions {
			adjustedCondition := condition
			for j := len(filterParams); j >= 1; j-- {
				oldParam := fmt.Sprintf("$%d", j)
				newParam := fmt.Sprintf("$%d", j+4)
				adjustedCondition = strings.Replace(adjustedCondition, oldParam, newParam, 1)
			}
			adjustedFilterConditions[i] = adjustedCondition
		}
		allConditions = append(coordinateConditions, adjustedFilterConditions...)
	}

	whereClause := r.buildWhereClause(allConditions)

	dataQuery := fmt.Sprintf(`
		SELECT `+utils.ApartmentWithConditionAndOwnerSelectFields+`
		%s %s
		ORDER BY a.created_at DESC
	`, baseQuery, whereClause)

	rows, err := r.db.Query(dataQuery, allParams...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartments", "query")
	}
	defer utils.CloseRows(rows)

	apartments, err := utils.ScanApartmentsWithConditionAndOwner(rows)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartments", "scan")
	}

	r.LoadApartmentsRelatedDataBatch(apartments)

	return apartments, nil
}

func (r *ApartmentRepository) GetStatusStatistics() (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM apartments
		GROUP BY status`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}
	defer rows.Close()

	statistics := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status statistics: %w", err)
		}
		statistics[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return statistics, nil
}

func (r *ApartmentRepository) GetCityStatistics() (map[string]int, error) {
	query := `
		SELECT c.name, COUNT(*) as count
		FROM apartments a
		JOIN cities c ON a.city_id = c.id
		GROUP BY c.name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get city statistics: %w", err)
	}
	defer rows.Close()

	statistics := make(map[string]int)
	for rows.Next() {
		var city string
		var count int
		if err := rows.Scan(&city, &count); err != nil {
			return nil, fmt.Errorf("failed to scan city statistics: %w", err)
		}
		statistics[city] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return statistics, nil
}

func (r *ApartmentRepository) GetDistrictStatistics() (map[string]int, error) {
	query := `
		SELECT d.name, COUNT(*) as count
		FROM apartments a
		JOIN districts d ON a.district_id = d.id
		GROUP BY d.name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get district statistics: %w", err)
	}
	defer rows.Close()

	statistics := make(map[string]int)
	for rows.Next() {
		var district string
		var count int
		if err := rows.Scan(&district, &count); err != nil {
			return nil, fmt.Errorf("failed to scan district statistics: %w", err)
		}
		statistics[district] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return statistics, nil
}

func (r *ApartmentRepository) CheckSmartAvailability(apartmentID int) (bool, error) {
	nowUTC := utils.GetCurrentTimeUTC()
	nowLocal := utils.ConvertOutputFromUTC(nowUTC)

	nextHour := time.Date(
		nowLocal.Year(), nowLocal.Month(), nowLocal.Day(),
		nowLocal.Hour()+1, 0, 0, 0, utils.KazakhstanTZ,
	)

	startTimeUTC := nextHour.UTC()
	endTimeUTC := startTimeUTC.Add(3 * time.Hour)

	query := `
		SELECT COUNT(*) 
		FROM bookings 
		WHERE apartment_id = $1 
		AND status IN ('pending', 'approved', 'active')
		AND (
			-- Проверяем пересечение начала нового бронирования с существующим (включая время уборки)
			(start_date <= $2 AND (end_date + INTERVAL '1 minute' * cleaning_duration) > $2) OR
			-- Проверяем пересечение конца нового бронирования с началом существующего
			(start_date < $3 AND end_date >= $3) OR  
			-- Проверяем если новое бронирование полностью покрывает существующее
			(start_date >= $2 AND end_date <= $3)
		)`

	var count int
	err := r.db.QueryRow(query, apartmentID, startTimeUTC, endTimeUTC).Scan(&count)
	if err != nil {
		return false, utils.HandleSQLError(err, "apartment smart availability", "check")
	}

	return count == 0, nil
}

func (r *ApartmentRepository) LoadApartmentsRelatedDataBatch(apartments []*domain.Apartment) {
	if len(apartments) == 0 {
		return
	}

	apartmentIDs := make([]int, len(apartments))
	apartmentMap := make(map[int]*domain.Apartment, len(apartments))

	for i, apt := range apartments {
		apartmentIDs[i] = apt.ID
		apartmentMap[apt.ID] = apt
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		if houseRulesMap, err := r.GetHouseRulesByApartmentIDBatch(apartmentIDs); err == nil {
			for aptID, rules := range houseRulesMap {
				if apt, exists := apartmentMap[aptID]; exists {
					apt.HouseRules = rules
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		if amenitiesMap, err := r.GetAmenitiesByApartmentIDBatch(apartmentIDs); err == nil {
			for aptID, amenities := range amenitiesMap {
				if apt, exists := apartmentMap[aptID]; exists {
					apt.Amenities = amenities
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		if photosMap, err := r.GetPhotosByApartmentIDBatch(apartmentIDs); err == nil {
			for aptID, photos := range photosMap {
				if apt, exists := apartmentMap[aptID]; exists {
					apt.Photos = photos
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		if locationsMap, err := r.GetLocationsByApartmentIDBatch(apartmentIDs); err == nil {
			for aptID, location := range locationsMap {
				if apt, exists := apartmentMap[aptID]; exists {
					apt.Location = location
				}
			}
		}
	}()

	wg.Wait()
}

func (r *ApartmentRepository) GetHouseRulesByApartmentIDBatch(apartmentIDs []int) (map[int][]*domain.HouseRules, error) {
	if len(apartmentIDs) == 0 {
		return make(map[int][]*domain.HouseRules), nil
	}

	placeholders := make([]string, len(apartmentIDs))
	args := make([]interface{}, len(apartmentIDs))
	for i, id := range apartmentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `
		SELECT ahr.apartment_id, hr.id, hr.name, hr.description, hr.created_at, hr.updated_at
		FROM apartment_house_rules ahr
		JOIN house_rules hr ON ahr.house_rule_id = hr.id
		WHERE ahr.apartment_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY ahr.apartment_id, hr.id
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "house rules batch", "get")
	}
	defer utils.CloseRows(rows)

	result := make(map[int][]*domain.HouseRules)
	for rows.Next() {
		var apartmentID int
		var rule domain.HouseRules

		err := rows.Scan(
			&apartmentID, &rule.ID, &rule.Name, &rule.Description,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "house rules batch", "scan")
		}

		result[apartmentID] = append(result[apartmentID], &rule)
	}

	return result, nil
}

func (r *ApartmentRepository) GetAmenitiesByApartmentIDBatch(apartmentIDs []int) (map[int][]*domain.PopularAmenities, error) {
	if len(apartmentIDs) == 0 {
		return make(map[int][]*domain.PopularAmenities), nil
	}

	placeholders := make([]string, len(apartmentIDs))
	args := make([]interface{}, len(apartmentIDs))
	for i, id := range apartmentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `
		SELECT aa.apartment_id, a.id, a.name, a.description, a.created_at, a.updated_at
		FROM apartment_amenities aa
		JOIN amenities a ON aa.amenity_id = a.id
		WHERE aa.apartment_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY aa.apartment_id, a.id
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "amenities batch", "get")
	}
	defer utils.CloseRows(rows)

	result := make(map[int][]*domain.PopularAmenities)
	for rows.Next() {
		var apartmentID int
		var amenity domain.PopularAmenities

		err := rows.Scan(
			&apartmentID, &amenity.ID, &amenity.Name, &amenity.Description,
			&amenity.CreatedAt, &amenity.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "amenities batch", "scan")
		}

		result[apartmentID] = append(result[apartmentID], &amenity)
	}

	return result, nil
}

func (r *ApartmentRepository) GetPhotosByApartmentIDBatch(apartmentIDs []int) (map[int][]*domain.ApartmentPhoto, error) {
	if len(apartmentIDs) == 0 {
		return make(map[int][]*domain.ApartmentPhoto), nil
	}

	placeholders := make([]string, len(apartmentIDs))
	args := make([]interface{}, len(apartmentIDs))
	for i, id := range apartmentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `
		SELECT apartment_id, id, url, "order", created_at, updated_at
		FROM apartment_photos
		WHERE apartment_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY apartment_id, "order", created_at
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "photos batch", "get")
	}
	defer utils.CloseRows(rows)

	result := make(map[int][]*domain.ApartmentPhoto)
	for rows.Next() {
		var photo domain.ApartmentPhoto

		err := rows.Scan(
			&photo.ApartmentID, &photo.ID, &photo.URL, &photo.Order,
			&photo.CreatedAt, &photo.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "photos batch", "scan")
		}

		result[photo.ApartmentID] = append(result[photo.ApartmentID], &photo)
	}

	return result, nil
}

func (r *ApartmentRepository) GetLocationsByApartmentIDBatch(apartmentIDs []int) (map[int]*domain.ApartmentLocation, error) {
	if len(apartmentIDs) == 0 {
		return make(map[int]*domain.ApartmentLocation), nil
	}

	placeholders := make([]string, len(apartmentIDs))
	args := make([]interface{}, len(apartmentIDs))
	for i, id := range apartmentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `
		SELECT id, apartment_id, latitude, longitude, created_at, updated_at
		FROM apartment_locations
		WHERE apartment_id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "locations batch", "get")
	}
	defer utils.CloseRows(rows)

	result := make(map[int]*domain.ApartmentLocation)
	for rows.Next() {
		var location domain.ApartmentLocation

		err := rows.Scan(
			&location.ID, &location.ApartmentID, &location.Latitude, &location.Longitude,
			&location.CreatedAt, &location.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "locations batch", "scan")
		}

		result[location.ApartmentID] = &location
	}

	return result, nil
}

func (r *ApartmentRepository) IncrementViewCount(apartmentID int) error {
	query := `UPDATE apartments SET view_count = view_count + 1 WHERE id = $1`

	_, err := r.db.Exec(query, apartmentID)
	if err != nil {
		return utils.HandleSQLError(err, "apartment view count", "increment")
	}

	return nil
}

func (r *ApartmentRepository) IncrementBookingCount(apartmentID int) error {
	query := `UPDATE apartments SET booking_count = booking_count + 1 WHERE id = $1`

	_, err := r.db.Exec(query, apartmentID)
	if err != nil {
		return utils.HandleSQLError(err, "apartment booking count", "increment")
	}

	return nil
}

func (r *ApartmentRepository) AdminUpdateViewCount(apartmentID int, viewCount int) error {
	query := `UPDATE apartments SET view_count = $2 WHERE id = $1`

	_, err := r.db.Exec(query, apartmentID, viewCount)
	if err != nil {
		return utils.HandleSQLError(err, "apartment view count", "admin update")
	}

	return nil
}

func (r *ApartmentRepository) AdminUpdateBookingCount(apartmentID int, bookingCount int) error {
	query := `UPDATE apartments SET booking_count = $2 WHERE id = $1`

	_, err := r.db.Exec(query, apartmentID, bookingCount)
	if err != nil {
		return utils.HandleSQLError(err, "apartment booking count", "admin update")
	}

	return nil
}

func (r *ApartmentRepository) AdminResetCounters(apartmentID int) error {
	query := `UPDATE apartments SET view_count = 0, booking_count = 0 WHERE id = $1`

	_, err := r.db.Exec(query, apartmentID)
	if err != nil {
		return utils.HandleSQLError(err, "apartment counters", "admin reset")
	}

	return nil
}
