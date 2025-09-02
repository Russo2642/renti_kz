package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/russo2642/renti_kz/internal/domain"
)

type conciergeRepository struct {
	db *sql.DB
}

func NewConciergeRepository(db *sql.DB) domain.ConciergeRepository {
	return &conciergeRepository{
		db: db,
	}
}

func (r *conciergeRepository) Create(concierge *domain.Concierge) error {
	query := `
		INSERT INTO concierges (user_id, is_active, schedule, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		concierge.UserID,
		concierge.IsActive,
		concierge.Schedule,
	).Scan(&concierge.ID, &concierge.CreatedAt, &concierge.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create concierge: %w", err)
	}

	return nil
}

func (r *conciergeRepository) GetByID(id int) (*domain.Concierge, error) {
	query := `
		SELECT c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			   u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
			   u.role_id, u.is_active, u.password_hash, u.created_at, u.updated_at
		FROM concierges c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.id = $1`

	concierge := &domain.Concierge{}
	user := &domain.User{}

	err := r.db.QueryRow(query, id).Scan(
		&concierge.ID, &concierge.UserID, &concierge.IsActive,
		&concierge.Schedule, &concierge.CreatedAt, &concierge.UpdatedAt,
		&user.ID, &user.Phone, &user.FirstName, &user.LastName, &user.Email,
		&user.CityID, &user.IIN, &user.RoleID, &user.IsActive, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("concierge not found")
		}
		return nil, fmt.Errorf("failed to get concierge: %w", err)
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

	concierge.User = user

	apartments, err := r.loadConciergeApartments(concierge.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load concierge apartments: %w", err)
	}
	concierge.Apartments = apartments

	return concierge, nil
}

func (r *conciergeRepository) GetByUserID(userID int) (*domain.Concierge, error) {
	query := `
		SELECT c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at
		FROM concierges c
		WHERE c.user_id = $1`

	concierge := &domain.Concierge{}

	err := r.db.QueryRow(query, userID).Scan(
		&concierge.ID, &concierge.UserID, &concierge.IsActive,
		&concierge.Schedule, &concierge.CreatedAt, &concierge.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("concierge not found")
		}
		return nil, fmt.Errorf("failed to get concierge by user ID: %w", err)
	}

	apartments, err := r.loadConciergeApartments(concierge.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load concierge apartments: %w", err)
	}
	concierge.Apartments = apartments

	return concierge, nil
}

func (r *conciergeRepository) GetByApartmentID(apartmentID int) ([]*domain.Concierge, error) {
	query := `
		SELECT c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			   u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
			   u.role_id, u.is_active, u.password_hash, u.created_at, u.updated_at
		FROM concierges c
		LEFT JOIN users u ON c.user_id = u.id
		INNER JOIN concierge_apartments ca ON c.id = ca.concierge_id
		WHERE ca.apartment_id = $1 AND ca.is_active = true`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get concierges by apartment ID: %w", err)
	}
	defer rows.Close()

	var concierges []*domain.Concierge

	for rows.Next() {
		concierge := &domain.Concierge{}
		user := &domain.User{}

		err := rows.Scan(
			&concierge.ID, &concierge.UserID, &concierge.IsActive,
			&concierge.Schedule, &concierge.CreatedAt, &concierge.UpdatedAt,
			&user.ID, &user.Phone, &user.FirstName, &user.LastName, &user.Email,
			&user.CityID, &user.IIN, &user.RoleID, &user.IsActive, &user.PasswordHash,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan concierge: %w", err)
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

		concierge.User = user

		apartments, err := r.loadConciergeApartments(concierge.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load concierge apartments: %w", err)
		}
		concierge.Apartments = apartments

		concierges = append(concierges, concierge)
	}

	return concierges, nil
}

func (r *conciergeRepository) GetByApartmentIDActive(apartmentID int) ([]*domain.Concierge, error) {
	query := `
		SELECT c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			   u.id, u.first_name, u.last_name, u.email, u.phone, u.created_at, u.updated_at
		FROM concierges c
		LEFT JOIN users u ON c.user_id = u.id
		INNER JOIN concierge_apartments ca ON c.id = ca.concierge_id
		WHERE ca.apartment_id = $1 AND ca.is_active = true AND c.is_active = true`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active concierges: %w", err)
	}
	defer rows.Close()

	var concierges []*domain.Concierge

	for rows.Next() {
		concierge := &domain.Concierge{}
		user := &domain.User{}

		err := rows.Scan(
			&concierge.ID, &concierge.UserID, &concierge.IsActive,
			&concierge.Schedule, &concierge.CreatedAt, &concierge.UpdatedAt,
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan concierge: %w", err)
		}

		concierge.User = user

		apartments, err := r.loadConciergeApartments(concierge.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load concierge apartments: %w", err)
		}
		concierge.Apartments = apartments

		concierges = append(concierges, concierge)
	}

	return concierges, nil
}

func (r *conciergeRepository) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.Concierge, int, error) {
	baseQuery := `
		FROM concierges c
		LEFT JOIN users u ON c.user_id = u.id`

	var whereConditions []string
	var args []interface{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "is_active":
			whereConditions = append(whereConditions, fmt.Sprintf("c.is_active = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "apartment_id":
			whereConditions = append(whereConditions, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM concierge_apartments ca 
				WHERE ca.concierge_id = c.id AND ca.apartment_id = $%d AND ca.is_active = true
			)`, argIndex))
			args = append(args, value)
			argIndex++
		case "user_id":
			whereConditions = append(whereConditions, fmt.Sprintf("c.user_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) " + baseQuery + " " + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count concierges: %w", err)
	}

	query := `
		SELECT c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			   u.id, u.first_name, u.last_name, u.email, u.phone, u.iin, u.city_id, u.role_id, u.created_at, u.updated_at
		` + baseQuery + " " + whereClause + `
		ORDER BY c.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get concierges: %w", err)
	}
	defer rows.Close()

	var concierges []*domain.Concierge

	for rows.Next() {
		concierge := &domain.Concierge{}
		user := &domain.User{}

		err := rows.Scan(
			&concierge.ID, &concierge.UserID, &concierge.IsActive,
			&concierge.Schedule, &concierge.CreatedAt, &concierge.UpdatedAt,
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
			&user.IIN, &user.CityID, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan concierge: %w", err)
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

		concierge.User = user

		apartments, err := r.loadConciergeApartments(concierge.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load concierge apartments: %w", err)
		}
		concierge.Apartments = apartments

		concierges = append(concierges, concierge)
	}

	return concierges, total, nil
}

func (r *conciergeRepository) Update(concierge *domain.Concierge) error {
	query := `
		UPDATE concierges 
		SET is_active = $2, schedule = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRow(
		query,
		concierge.ID,
		concierge.IsActive,
		concierge.Schedule,
	).Scan(&concierge.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update concierge: %w", err)
	}

	return nil
}

func (r *conciergeRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM concierge_apartments WHERE concierge_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete concierge apartments: %w", err)
	}

	query := `DELETE FROM concierges WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete concierge: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("concierge not found")
	}

	return nil
}

func (r *conciergeRepository) IsUserConcierge(userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM concierges WHERE user_id = $1 AND is_active = TRUE)`

	var exists bool
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is concierge: %w", err)
	}

	return exists, nil
}

func (r *conciergeRepository) GetConciergesByOwner(ownerID int) ([]*domain.Concierge, error) {
	query := `
		SELECT DISTINCT c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			   u.id, u.first_name, u.last_name, u.email, u.phone, u.created_at, u.updated_at
		FROM concierges c
		LEFT JOIN users u ON c.user_id = u.id
		INNER JOIN concierge_apartments ca ON c.id = ca.concierge_id
		INNER JOIN apartments a ON ca.apartment_id = a.id
		LEFT JOIN property_owners po ON a.id = po.apartment_id
		WHERE po.user_id = $1 AND c.is_active = TRUE AND ca.is_active = TRUE
		ORDER BY c.created_at DESC`

	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get concierges by owner: %w", err)
	}
	defer rows.Close()

	var concierges []*domain.Concierge
	conciergeMap := make(map[int]*domain.Concierge)

	for rows.Next() {
		concierge := &domain.Concierge{}
		user := &domain.User{}

		err := rows.Scan(
			&concierge.ID, &concierge.UserID, &concierge.IsActive,
			&concierge.Schedule, &concierge.CreatedAt, &concierge.UpdatedAt,
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan concierge: %w", err)
		}

		if _, exists := conciergeMap[concierge.ID]; exists {
			continue
		}

		concierge.User = user

		apartments, err := r.loadConciergeApartments(concierge.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load concierge apartments: %w", err)
		}
		concierge.Apartments = apartments

		conciergeMap[concierge.ID] = concierge
		concierges = append(concierges, concierge)
	}

	return concierges, nil
}

func (r *conciergeRepository) AssignToApartment(conciergeID, apartmentID int) error {

	updateQuery := `
		UPDATE concierge_apartments 
		SET is_active = true, assigned_at = NOW(), updated_at = NOW()
		WHERE concierge_id = $1 AND apartment_id = $2 AND is_active = false`

	result, err := r.db.Exec(updateQuery, conciergeID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to update concierge apartment assignment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO concierge_apartments (concierge_id, apartment_id, is_active, assigned_at, created_at, updated_at)
		VALUES ($1, $2, true, NOW(), NOW(), NOW())
		ON CONFLICT (concierge_id, apartment_id) DO NOTHING`

	_, err = r.db.Exec(insertQuery, conciergeID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to assign concierge to apartment: %w", err)
	}

	return nil
}

func (r *conciergeRepository) RemoveFromApartment(conciergeID, apartmentID int) error {
	query := `
		UPDATE concierge_apartments 
		SET is_active = false, updated_at = NOW()
		WHERE concierge_id = $1 AND apartment_id = $2 AND is_active = true`

	result, err := r.db.Exec(query, conciergeID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to remove concierge from apartment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("concierge assignment not found or already inactive")
	}

	return nil
}

func (r *conciergeRepository) GetConciergeApartments(conciergeID int) ([]*domain.ConciergeApartment, error) {
	query := `
		SELECT ca.id, ca.concierge_id, ca.apartment_id, ca.is_active, ca.assigned_at, ca.created_at, ca.updated_at,
			   a.id, a.description, a.street, a.building, a.city_id, a.price, a.is_free, a.created_at, a.updated_at
		FROM concierge_apartments ca
		LEFT JOIN apartments a ON ca.apartment_id = a.id
		WHERE ca.concierge_id = $1 AND ca.is_active = true
		ORDER BY ca.assigned_at DESC`

	rows, err := r.db.Query(query, conciergeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get concierge apartments: %w", err)
	}
	defer rows.Close()

	var conciergeApartments []*domain.ConciergeApartment

	for rows.Next() {
		ca := &domain.ConciergeApartment{}
		apartment := &domain.Apartment{}

		err := rows.Scan(
			&ca.ID, &ca.ConciergeID, &ca.ApartmentID, &ca.IsActive, &ca.AssignedAt, &ca.CreatedAt, &ca.UpdatedAt,
			&apartment.ID, &apartment.Description, &apartment.Street, &apartment.Building,
			&apartment.CityID, &apartment.Price, &apartment.IsFree,
			&apartment.CreatedAt, &apartment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan concierge apartment: %w", err)
		}

		ca.Apartment = apartment
		conciergeApartments = append(conciergeApartments, ca)
	}

	return conciergeApartments, nil
}

func (r *conciergeRepository) GetApartmentConcierges(apartmentID int) ([]*domain.ConciergeApartment, error) {
	query := `
		SELECT ca.id, ca.concierge_id, ca.apartment_id, ca.is_active, ca.assigned_at, ca.created_at, ca.updated_at
		FROM concierge_apartments ca
		WHERE ca.apartment_id = $1 AND ca.is_active = true
		ORDER BY ca.assigned_at DESC`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment concierges: %w", err)
	}
	defer rows.Close()

	var conciergeApartments []*domain.ConciergeApartment

	for rows.Next() {
		ca := &domain.ConciergeApartment{}

		err := rows.Scan(
			&ca.ID, &ca.ConciergeID, &ca.ApartmentID, &ca.IsActive, &ca.AssignedAt, &ca.CreatedAt, &ca.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan concierge apartment: %w", err)
		}

		conciergeApartments = append(conciergeApartments, ca)
	}

	return conciergeApartments, nil
}

func (r *conciergeRepository) loadConciergeApartments(conciergeID int) ([]*domain.Apartment, error) {
	query := `
		SELECT a.id, a.owner_id, a.city_id, a.district_id, a.microdistrict_id,
			   a.street, a.building, a.apartment_number, a.residential_complex, a.room_count,
			   a.total_area, a.kitchen_area, a.floor, a.total_floors, a.condition_id,
			   a.price, a.daily_price, COALESCE(a.service_fee_percentage, 0), a.rental_type_hourly, a.rental_type_daily,
			   a.is_free, a.status, a.moderator_comment, a.description, a.listing_type,
			   a.is_agreement_accepted, a.agreement_accepted_at, a.contract_id,
			   a.created_at, a.updated_at
		FROM apartments a
		INNER JOIN concierge_apartments ca ON a.id = ca.apartment_id
		WHERE ca.concierge_id = $1 AND ca.is_active = true
		ORDER BY ca.assigned_at DESC`

	rows, err := r.db.Query(query, conciergeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load apartments: %w", err)
	}
	defer rows.Close()

	var apartments []*domain.Apartment

	for rows.Next() {
		apartment := &domain.Apartment{}

		err := rows.Scan(
			&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID, &apartment.MicrodistrictID,
			&apartment.Street, &apartment.Building, &apartment.ApartmentNumber, &apartment.ResidentialComplex, &apartment.RoomCount,
			&apartment.TotalArea, &apartment.KitchenArea, &apartment.Floor, &apartment.TotalFloors, &apartment.ConditionID,
			&apartment.Price, &apartment.DailyPrice, &apartment.ServiceFeePercentage, &apartment.RentalTypeHourly, &apartment.RentalTypeDaily,
			&apartment.IsFree, &apartment.Status, &apartment.ModeratorComment, &apartment.Description, &apartment.ListingType,
			&apartment.IsAgreementAccepted, &apartment.AgreementAcceptedAt, &apartment.ContractID,
			&apartment.CreatedAt, &apartment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan apartment: %w", err)
		}

		apartments = append(apartments, apartment)
	}

	return apartments, nil
}
