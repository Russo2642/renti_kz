package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type CleanerRepository struct {
	db *sql.DB
}

func NewCleanerRepository(db *sql.DB) *CleanerRepository {
	return &CleanerRepository{
		db: db,
	}
}

func (r *CleanerRepository) Create(cleaner *domain.Cleaner) error {
	query := `
		INSERT INTO cleaners (user_id, is_active, schedule, created_at, updated_at) 
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		cleaner.UserID,
		cleaner.IsActive,
		cleaner.Schedule,
	).Scan(&cleaner.ID, &cleaner.CreatedAt, &cleaner.UpdatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "cleaner", "create")
	}

	return nil
}

func (r *CleanerRepository) GetByID(id int) (*domain.Cleaner, error) {
	query := `
		SELECT 
			c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM cleaners c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.id = $1`

	cleaner := &domain.Cleaner{}
	user := &domain.User{}
	var userID, userRoleID, userCityID sql.NullInt32
	var userPhone, userFirstName, userLastName, userEmail, userIIN sql.NullString
	var userIsActive sql.NullBool
	var userCreatedAt, userUpdatedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&cleaner.ID, &cleaner.UserID, &cleaner.IsActive, &cleaner.Schedule, &cleaner.CreatedAt, &cleaner.UpdatedAt,
		&userID, &userPhone, &userFirstName, &userLastName, &userEmail, &userCityID, &userIIN, &userRoleID, &userIsActive, &userCreatedAt, &userUpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "cleaner", "get by id")
	}

	if userID.Valid {
		user.ID = int(userID.Int32)
		user.Phone = userPhone.String
		user.FirstName = userFirstName.String
		user.LastName = userLastName.String
		user.Email = userEmail.String
		user.CityID = int(userCityID.Int32)
		user.IIN = userIIN.String
		user.RoleID = int(userRoleID.Int32)
		user.IsActive = userIsActive.Bool
		user.CreatedAt = userCreatedAt.Time
		user.UpdatedAt = userUpdatedAt.Time
		cleaner.User = user
	}

	// Загружаем квартиры уборщицы
	apartments, err := r.loadCleanerApartments(cleaner.ID)
	if err != nil {
		log.Printf("Ошибка загрузки квартир для уборщицы %d: %v", cleaner.ID, err)
	} else {
		cleaner.Apartments = apartments
	}

	return cleaner, nil
}

func (r *CleanerRepository) GetByUserID(userID int) (*domain.Cleaner, error) {
	query := `
		SELECT 
			c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at
		FROM cleaners c
		WHERE c.user_id = $1`

	cleaner := &domain.Cleaner{}

	err := r.db.QueryRow(query, userID).Scan(
		&cleaner.ID, &cleaner.UserID, &cleaner.IsActive, &cleaner.Schedule, &cleaner.CreatedAt, &cleaner.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "cleaner", "get by user id")
	}

	return cleaner, nil
}

func (r *CleanerRepository) GetByApartmentID(apartmentID int) ([]*domain.Cleaner, error) {
	query := `
		SELECT 
			c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM cleaners c
		INNER JOIN cleaner_apartments ca ON c.id = ca.cleaner_id
		LEFT JOIN users u ON c.user_id = u.id
		WHERE ca.apartment_id = $1 AND ca.is_active = TRUE`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "cleaners by apartment", "query")
	}
	defer utils.CloseRows(rows)

	return r.scanCleaners(rows)
}

func (r *CleanerRepository) GetByApartmentIDActive(apartmentID int) ([]*domain.Cleaner, error) {
	query := `
		SELECT 
			c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM cleaners c
		INNER JOIN cleaner_apartments ca ON c.id = ca.cleaner_id
		LEFT JOIN users u ON c.user_id = u.id
		WHERE ca.apartment_id = $1 AND ca.is_active = TRUE AND c.is_active = TRUE`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "active cleaners by apartment", "query")
	}
	defer utils.CloseRows(rows)

	return r.scanCleaners(rows)
}

func (r *CleanerRepository) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.Cleaner, int, error) {
	conditions, params, paramIndex := r.buildFilterConditions(filters)
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + fmt.Sprintf("%v", conditions[0])
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + fmt.Sprintf("%v", conditions[i])
		}
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM cleaners c
		LEFT JOIN users u ON c.user_id = u.id
		%s`, whereClause)

	var total int
	err := r.db.QueryRow(countQuery, params...).Scan(&total)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "cleaners count", "query")
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT 
			c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM cleaners c
		LEFT JOIN users u ON c.user_id = u.id
		%s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, paramIndex, paramIndex+1)

	params = append(params, pageSize, offset)

	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "cleaners", "query")
	}
	defer utils.CloseRows(rows)

	cleaners, err := r.scanCleaners(rows)
	if err != nil {
		return nil, 0, err
	}

	return cleaners, total, nil
}

func (r *CleanerRepository) Update(cleaner *domain.Cleaner) error {
	query := `
		UPDATE cleaners SET 
			is_active = $2, schedule = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.Exec(
		query,
		cleaner.ID,
		cleaner.IsActive,
		cleaner.Schedule,
	)

	if err != nil {
		return utils.HandleSQLErrorWithID(err, "cleaner", "update", cleaner.ID)
	}

	return nil
}

func (r *CleanerRepository) UpdatePartial(id int, request *domain.UpdateCleanerRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	// Добавляем поля только если они переданы
	if request.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *request.IsActive)
		argIndex++
	}

	if request.Schedule != nil {
		// Для расписания делаем merge с существующим
		mergedSchedule, err := r.mergeSchedule(id, request.Schedule)
		if err != nil {
			return fmt.Errorf("ошибка объединения расписания: %w", err)
		}

		setParts = append(setParts, fmt.Sprintf("schedule = $%d", argIndex))
		args = append(args, mergedSchedule)
		argIndex++
	}

	// Если нет полей для обновления, ничего не делаем
	if len(setParts) == 0 {
		return nil
	}

	// Всегда обновляем updated_at
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	query := fmt.Sprintf(`
		UPDATE cleaners SET %s
		WHERE id = $%d`,
		strings.Join(setParts, ", "),
		argIndex,
	)

	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "cleaner", "partial update", id)
	}

	return nil
}

// Метод для объединения нового расписания с существующим
func (r *CleanerRepository) mergeSchedule(cleanerID int, newSchedule *domain.CleanerSchedule) (*domain.CleanerSchedule, error) {
	// Получаем текущее расписание
	query := `SELECT schedule FROM cleaners WHERE id = $1`
	var currentScheduleJSON sql.NullString
	err := r.db.QueryRow(query, cleanerID).Scan(&currentScheduleJSON)
	if err != nil {
		return nil, utils.HandleSQLError(err, "cleaner schedule", "get current")
	}

	// Создаем результирующее расписание
	var mergedSchedule *domain.CleanerSchedule

	// Если есть текущее расписание, парсим его
	if currentScheduleJSON.Valid && currentScheduleJSON.String != "" {
		currentSchedule := &domain.CleanerSchedule{}
		err = json.Unmarshal([]byte(currentScheduleJSON.String), currentSchedule)
		if err != nil {
			// Если не удалось распарсить, используем пустое расписание
			mergedSchedule = &domain.CleanerSchedule{}
		} else {
			mergedSchedule = currentSchedule
		}
	} else {
		// Если текущего расписания нет, создаем пустое
		mergedSchedule = &domain.CleanerSchedule{}
	}

	// Объединяем с новым расписанием (только непустые дни)
	if len(newSchedule.Monday) > 0 {
		mergedSchedule.Monday = newSchedule.Monday
	}
	if len(newSchedule.Tuesday) > 0 {
		mergedSchedule.Tuesday = newSchedule.Tuesday
	}
	if len(newSchedule.Wednesday) > 0 {
		mergedSchedule.Wednesday = newSchedule.Wednesday
	}
	if len(newSchedule.Thursday) > 0 {
		mergedSchedule.Thursday = newSchedule.Thursday
	}
	if len(newSchedule.Friday) > 0 {
		mergedSchedule.Friday = newSchedule.Friday
	}
	if len(newSchedule.Saturday) > 0 {
		mergedSchedule.Saturday = newSchedule.Saturday
	}
	if len(newSchedule.Sunday) > 0 {
		mergedSchedule.Sunday = newSchedule.Sunday
	}

	return mergedSchedule, nil
}

func (r *CleanerRepository) UpdateSchedulePatch(cleanerID int, schedulePatch *domain.CleanerSchedulePatch) error {

	query := `SELECT schedule FROM cleaners WHERE id = $1`
	var currentScheduleJSON sql.NullString
	err := r.db.QueryRow(query, cleanerID).Scan(&currentScheduleJSON)
	if err != nil {
		return utils.HandleSQLError(err, "cleaner schedule", "get current")
	}

	var mergedSchedule *domain.CleanerSchedule

	if currentScheduleJSON.Valid && currentScheduleJSON.String != "" {
		currentSchedule := &domain.CleanerSchedule{}
		err = json.Unmarshal([]byte(currentScheduleJSON.String), currentSchedule)
		if err != nil {
			mergedSchedule = &domain.CleanerSchedule{}
		} else {
			mergedSchedule = currentSchedule
		}
	} else {
		mergedSchedule = &domain.CleanerSchedule{}
	}

	if schedulePatch.Monday != nil {
		mergedSchedule.Monday = *schedulePatch.Monday
	}
	if schedulePatch.Tuesday != nil {
		mergedSchedule.Tuesday = *schedulePatch.Tuesday
	}
	if schedulePatch.Wednesday != nil {
		mergedSchedule.Wednesday = *schedulePatch.Wednesday
	}
	if schedulePatch.Thursday != nil {
		mergedSchedule.Thursday = *schedulePatch.Thursday
	}
	if schedulePatch.Friday != nil {
		mergedSchedule.Friday = *schedulePatch.Friday
	}
	if schedulePatch.Saturday != nil {
		mergedSchedule.Saturday = *schedulePatch.Saturday
	}
	if schedulePatch.Sunday != nil {
		mergedSchedule.Sunday = *schedulePatch.Sunday
	}

	updateQuery := `
		UPDATE cleaners 
		SET schedule = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`

	_, err = r.db.Exec(updateQuery, mergedSchedule, cleanerID)
	if err != nil {
		return utils.HandleSQLError(err, "cleaner schedule", "update patch")
	}

	return nil
}

func (r *CleanerRepository) Delete(id int) error {
	query := `DELETE FROM cleaners WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "cleaner", "delete", id)
	}
	return nil
}

func (r *CleanerRepository) IsUserCleaner(userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM cleaners WHERE user_id = $1 AND is_active = TRUE)`
	var exists bool
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, utils.HandleSQLError(err, "cleaner existence check", "query")
	}
	return exists, nil
}

func (r *CleanerRepository) GetCleanersByOwner(ownerID int) ([]*domain.Cleaner, error) {
	query := `
		SELECT DISTINCT
			c.id, c.user_id, c.is_active, c.schedule, c.created_at, c.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM cleaners c
		INNER JOIN cleaner_apartments ca ON c.id = ca.cleaner_id
		INNER JOIN apartments a ON ca.apartment_id = a.id
		LEFT JOIN users u ON c.user_id = u.id
		WHERE a.owner_id = $1 AND ca.is_active = TRUE`

	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "cleaners by owner", "query")
	}
	defer utils.CloseRows(rows)

	return r.scanCleaners(rows)
}

func (r *CleanerRepository) AssignToApartment(cleanerID, apartmentID int) error {
	// Сначала проверяем есть ли уже активная связь
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM cleaner_apartments WHERE cleaner_id = $1 AND apartment_id = $2 AND is_active = TRUE)`
	err := r.db.QueryRow(checkQuery, cleanerID, apartmentID).Scan(&exists)
	if err != nil {
		return utils.HandleSQLError(err, "cleaner apartment check", "query")
	}

	if exists {
		// Связь уже существует, просто обновляем время
		updateQuery := `UPDATE cleaner_apartments SET updated_at = CURRENT_TIMESTAMP WHERE cleaner_id = $1 AND apartment_id = $2 AND is_active = TRUE`
		_, err = r.db.Exec(updateQuery, cleanerID, apartmentID)
		if err != nil {
			return utils.HandleSQLError(err, "cleaner apartment assignment", "update")
		}
		return nil
	}

	// Сначала деактивируем старые связи для этой пары
	deactivateQuery := `UPDATE cleaner_apartments SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP WHERE cleaner_id = $1 AND apartment_id = $2 AND is_active = TRUE`
	_, err = r.db.Exec(deactivateQuery, cleanerID, apartmentID)
	if err != nil {
		return utils.HandleSQLError(err, "cleaner apartment deactivation", "update")
	}

	// Теперь создаем новую активную связь
	insertQuery := `
		INSERT INTO cleaner_apartments (cleaner_id, apartment_id, is_active, assigned_at, created_at, updated_at) 
		VALUES ($1, $2, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = r.db.Exec(insertQuery, cleanerID, apartmentID)
	if err != nil {
		return utils.HandleSQLError(err, "cleaner apartment assignment", "create")
	}

	return nil
}

func (r *CleanerRepository) RemoveFromApartment(cleanerID, apartmentID int) error {
	query := `
		UPDATE cleaner_apartments 
		SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE cleaner_id = $1 AND apartment_id = $2 AND is_active = TRUE`

	_, err := r.db.Exec(query, cleanerID, apartmentID)
	if err != nil {
		return utils.HandleSQLError(err, "cleaner apartment removal", "update")
	}

	return nil
}

func (r *CleanerRepository) GetCleanerApartments(cleanerID int) ([]*domain.CleanerApartment, error) {
	query := `
		SELECT 
			ca.id, ca.cleaner_id, ca.apartment_id, ca.is_active, ca.assigned_at, ca.created_at, ca.updated_at
		FROM cleaner_apartments ca
		WHERE ca.cleaner_id = $1 AND ca.is_active = TRUE
		ORDER BY ca.assigned_at DESC`

	rows, err := r.db.Query(query, cleanerID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "cleaner apartments", "query")
	}
	defer utils.CloseRows(rows)

	var cleanerApartments []*domain.CleanerApartment
	for rows.Next() {
		ca := &domain.CleanerApartment{}
		err := rows.Scan(
			&ca.ID, &ca.CleanerID, &ca.ApartmentID, &ca.IsActive, &ca.AssignedAt, &ca.CreatedAt, &ca.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "cleaner apartment", "scan")
		}
		cleanerApartments = append(cleanerApartments, ca)
	}

	if err = utils.CheckRowsError(rows, "cleaner apartments iteration"); err != nil {
		return nil, err
	}

	return cleanerApartments, nil
}

func (r *CleanerRepository) GetApartmentCleaners(apartmentID int) ([]*domain.CleanerApartment, error) {
	query := `
		SELECT 
			ca.id, ca.cleaner_id, ca.apartment_id, ca.is_active, ca.assigned_at, ca.created_at, ca.updated_at
		FROM cleaner_apartments ca
		WHERE ca.apartment_id = $1 AND ca.is_active = TRUE
		ORDER BY ca.assigned_at DESC`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment cleaners", "query")
	}
	defer utils.CloseRows(rows)

	var cleanerApartments []*domain.CleanerApartment
	for rows.Next() {
		ca := &domain.CleanerApartment{}
		err := rows.Scan(
			&ca.ID, &ca.CleanerID, &ca.ApartmentID, &ca.IsActive, &ca.AssignedAt, &ca.CreatedAt, &ca.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "cleaner apartment", "scan")
		}
		cleanerApartments = append(cleanerApartments, ca)
	}

	if err = utils.CheckRowsError(rows, "apartment cleaners iteration"); err != nil {
		return nil, err
	}

	return cleanerApartments, nil
}

func (r *CleanerRepository) GetApartmentsForCleaning(cleanerID int) ([]*domain.ApartmentForCleaning, error) {
	query := `
		SELECT 
			a.id, a.street, a.building, a.apartment_number, a.room_count, a.total_area, a.is_free,
			lb.end_date as last_booking_end_date,
			c.name as city_name,
			d.name as district_name
		FROM apartments a
		INNER JOIN cleaner_apartments ca ON a.id = ca.apartment_id
		LEFT JOIN cities c ON a.city_id = c.id
		LEFT JOIN districts d ON a.district_id = d.id
		LEFT JOIN (
			SELECT DISTINCT ON (apartment_id) 
				apartment_id, end_date
			FROM bookings 
			WHERE status = 'completed'
			ORDER BY apartment_id, end_date DESC
		) lb ON a.id = lb.apartment_id
		WHERE ca.cleaner_id = $1 
		AND ca.is_active = TRUE
		AND a.is_free = FALSE  -- Квартира помечена как занятая
		AND NOT EXISTS (
			SELECT 1 FROM bookings b 
			WHERE b.apartment_id = a.id 
			AND (
				-- Активные прямо сейчас
				(b.status = 'active' AND b.start_date <= NOW() AND b.end_date > NOW())
				OR
				-- Подтвержденные/ожидающие в ближайшие 2 часа
				(b.status IN ('approved', 'pending', 'awaiting_payment') 
				 AND b.start_date BETWEEN NOW() AND NOW() + INTERVAL '2 hours')
				OR
				-- Недавно созданные бронирования с началом в ближайшие 2 часа
				(b.status = 'created' 
				 AND b.created_at > NOW() - INTERVAL '30 minutes'
				 AND b.start_date BETWEEN NOW() AND NOW() + INTERVAL '2 hours')
			)
		)
		AND EXISTS (
			SELECT 1 FROM bookings b2 
			WHERE b2.apartment_id = a.id 
			AND b2.status = 'completed'
			AND b2.end_date > NOW() - INTERVAL '24 hours'  -- За последние 24 часа
		)
		ORDER BY lb.end_date ASC NULLS LAST`

	rows, err := r.db.Query(query, cleanerID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartments for cleaning", "query")
	}
	defer utils.CloseRows(rows)

	var apartments []*domain.ApartmentForCleaning
	for rows.Next() {
		apt := &domain.ApartmentForCleaning{}
		var lastBookingEndDate sql.NullTime
		var cityName, districtName sql.NullString

		err := rows.Scan(
			&apt.ID, &apt.Street, &apt.Building, &apt.ApartmentNumber, &apt.RoomCount, &apt.TotalArea, &apt.IsFree,
			&lastBookingEndDate, &cityName, &districtName,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment for cleaning", "scan")
		}

		if lastBookingEndDate.Valid {
			apt.LastBookingEndDate = &lastBookingEndDate.Time
			timeSince := time.Since(lastBookingEndDate.Time)
			if timeSince < time.Hour {
				timeSinceStr := fmt.Sprintf("%d мин назад", int(timeSince.Minutes()))
				apt.TimeSinceLastBooking = &timeSinceStr
			} else if timeSince < 24*time.Hour {
				timeSinceStr := fmt.Sprintf("%d ч назад", int(timeSince.Hours()))
				apt.TimeSinceLastBooking = &timeSinceStr
			} else {
				timeSinceStr := fmt.Sprintf("%d дн назад", int(timeSince.Hours()/24))
				apt.TimeSinceLastBooking = &timeSinceStr
			}
		}

		apt.CleaningStatus = "needs_cleaning"

		if cityName.Valid {
			apt.City = &domain.City{Name: cityName.String}
		}
		if districtName.Valid {
			apt.District = &domain.District{Name: districtName.String}
		}

		apartments = append(apartments, apt)
	}

	if err = utils.CheckRowsError(rows, "apartments for cleaning iteration"); err != nil {
		return nil, err
	}

	return apartments, nil
}

func (r *CleanerRepository) GetApartmentsNeedingCleaning() ([]*domain.ApartmentForCleaning, error) {
	query := `
		SELECT 
			a.id, a.street, a.building, a.apartment_number, a.room_count, a.total_area, a.is_free,
			lb.end_date as last_booking_end_date,
			c.name as city_name,
			d.name as district_name
		FROM apartments a
		LEFT JOIN cities c ON a.city_id = c.id
		LEFT JOIN districts d ON a.district_id = d.id
		LEFT JOIN (
			SELECT DISTINCT ON (apartment_id) 
				apartment_id, end_date
			FROM bookings 
			WHERE status = 'completed'
			ORDER BY apartment_id, end_date DESC
		) lb ON a.id = lb.apartment_id
		WHERE a.is_free = FALSE  -- Квартира помечена как занятая
		AND NOT EXISTS (
			SELECT 1 FROM bookings b 
			WHERE b.apartment_id = a.id 
			AND b.status = 'active'  -- Но нет активных бронирований
			AND b.start_date <= NOW() 
			AND b.end_date > NOW()
		)
		AND EXISTS (
			SELECT 1 FROM bookings b2 
			WHERE b2.apartment_id = a.id 
			AND b2.status = 'completed'
			AND b2.end_date > NOW() - INTERVAL '24 hours'  -- За последние 24 часа
		)
		ORDER BY lb.end_date ASC NULLS LAST`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartments needing cleaning", "query")
	}
	defer utils.CloseRows(rows)

	var apartments []*domain.ApartmentForCleaning
	for rows.Next() {
		apt := &domain.ApartmentForCleaning{}
		var lastBookingEndDate sql.NullTime
		var cityName, districtName sql.NullString

		err := rows.Scan(
			&apt.ID, &apt.Street, &apt.Building, &apt.ApartmentNumber, &apt.RoomCount, &apt.TotalArea, &apt.IsFree,
			&lastBookingEndDate, &cityName, &districtName,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment needing cleaning", "scan")
		}

		if lastBookingEndDate.Valid {
			apt.LastBookingEndDate = &lastBookingEndDate.Time
			timeSince := time.Since(lastBookingEndDate.Time)
			if timeSince < time.Hour {
				timeSinceStr := fmt.Sprintf("%d мин назад", int(timeSince.Minutes()))
				apt.TimeSinceLastBooking = &timeSinceStr
			} else if timeSince < 24*time.Hour {
				timeSinceStr := fmt.Sprintf("%d ч назад", int(timeSince.Hours()))
				apt.TimeSinceLastBooking = &timeSinceStr
			} else {
				timeSinceStr := fmt.Sprintf("%d дн назад", int(timeSince.Hours()/24))
				apt.TimeSinceLastBooking = &timeSinceStr
			}
		}

		apt.CleaningStatus = "needs_cleaning"

		if cityName.Valid {
			apt.City = &domain.City{Name: cityName.String}
		}
		if districtName.Valid {
			apt.District = &domain.District{Name: districtName.String}
		}

		apartments = append(apartments, apt)
	}

	if err = utils.CheckRowsError(rows, "apartments needing cleaning iteration"); err != nil {
		return nil, err
	}

	return apartments, nil
}

func (r *CleanerRepository) buildFilterConditions(filters map[string]interface{}) ([]string, []interface{}, int) {
	var conditions []string
	var params []interface{}
	paramIndex := 1

	for key, value := range filters {
		switch key {
		case "is_active":
			conditions = append(conditions, fmt.Sprintf("c.is_active = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "user_id":
			conditions = append(conditions, fmt.Sprintf("c.user_id = $%d", paramIndex))
			params = append(params, value)
			paramIndex++
		case "apartment_id":
			conditions = append(conditions, fmt.Sprintf("EXISTS(SELECT 1 FROM cleaner_apartments ca WHERE ca.cleaner_id = c.id AND ca.apartment_id = $%d AND ca.is_active = TRUE)", paramIndex))
			params = append(params, value)
			paramIndex++
		}
	}

	return conditions, params, paramIndex
}

func (r *CleanerRepository) scanCleaners(rows *sql.Rows) ([]*domain.Cleaner, error) {
	var cleaners []*domain.Cleaner
	for rows.Next() {
		cleaner := &domain.Cleaner{}
		user := &domain.User{}
		var userID, userRoleID, userCityID sql.NullInt32
		var userPhone, userFirstName, userLastName, userEmail, userIIN sql.NullString
		var userIsActive sql.NullBool
		var userCreatedAt, userUpdatedAt sql.NullTime

		err := rows.Scan(
			&cleaner.ID, &cleaner.UserID, &cleaner.IsActive, &cleaner.Schedule, &cleaner.CreatedAt, &cleaner.UpdatedAt,
			&userID, &userPhone, &userFirstName, &userLastName, &userEmail, &userCityID, &userIIN, &userRoleID, &userIsActive, &userCreatedAt, &userUpdatedAt,
		)

		if err != nil {
			return nil, utils.HandleSQLError(err, "cleaner", "scan")
		}

		if userID.Valid {
			user.ID = int(userID.Int32)
			user.Phone = userPhone.String
			user.FirstName = userFirstName.String
			user.LastName = userLastName.String
			user.Email = userEmail.String
			user.CityID = int(userCityID.Int32)
			user.IIN = userIIN.String
			user.RoleID = int(userRoleID.Int32)
			user.IsActive = userIsActive.Bool
			user.CreatedAt = userCreatedAt.Time
			user.UpdatedAt = userUpdatedAt.Time
			cleaner.User = user
		}

		cleaners = append(cleaners, cleaner)
	}

	if err := utils.CheckRowsError(rows, "cleaners iteration"); err != nil {
		return nil, err
	}

	// Загружаем квартиры для каждой уборщицы
	for _, cleaner := range cleaners {
		apartments, err := r.loadCleanerApartments(cleaner.ID)
		if err != nil {
			log.Printf("Ошибка загрузки квартир для уборщицы %d: %v", cleaner.ID, err)
			continue
		}
		cleaner.Apartments = apartments
	}

	return cleaners, nil
}

// Вспомогательный метод для загрузки квартир уборщицы
func (r *CleanerRepository) loadCleanerApartments(cleanerID int) ([]*domain.Apartment, error) {
	query := `
		SELECT 
			a.id, a.owner_id, a.city_id, a.district_id, a.street, a.building, 
			a.apartment_number, a.room_count, a.total_area, a.kitchen_area, 
			a.floor, a.total_floors, a.condition_id, a.price, a.daily_price, 
			a.service_fee_percentage, a.rental_type_hourly, a.rental_type_daily, 
			a.is_free, a.status, a.description, a.listing_type, 
			a.is_agreement_accepted, a.view_count, a.booking_count, 
			a.created_at, a.updated_at,
			c.id as city_id_full, c.name as city_name, c.region_id, c.created_at as city_created_at, c.updated_at as city_updated_at,
			d.id as district_id_full, d.name as district_name, d.city_id as district_city_id, d.created_at as district_created_at, d.updated_at as district_updated_at
		FROM apartments a
		INNER JOIN cleaner_apartments ca ON a.id = ca.apartment_id
		LEFT JOIN cities c ON a.city_id = c.id
		LEFT JOIN districts d ON a.district_id = d.id
		WHERE ca.cleaner_id = $1 AND ca.is_active = TRUE
		ORDER BY ca.assigned_at DESC`

	rows, err := r.db.Query(query, cleanerID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "cleaner apartments", "query")
	}
	defer utils.CloseRows(rows)

	var apartments []*domain.Apartment
	for rows.Next() {
		apartment := &domain.Apartment{}
		city := &domain.City{}
		district := &domain.District{}

		var cityID, districtID, cityRegionID, districtCityID sql.NullInt32
		var cityName, districtName sql.NullString
		var cityCreatedAt, cityUpdatedAt, districtCreatedAt, districtUpdatedAt sql.NullTime

		err := rows.Scan(
			&apartment.ID, &apartment.OwnerID, &apartment.CityID, &apartment.DistrictID,
			&apartment.Street, &apartment.Building, &apartment.ApartmentNumber,
			&apartment.RoomCount, &apartment.TotalArea, &apartment.KitchenArea,
			&apartment.Floor, &apartment.TotalFloors, &apartment.ConditionID,
			&apartment.Price, &apartment.DailyPrice, &apartment.ServiceFeePercentage,
			&apartment.RentalTypeHourly, &apartment.RentalTypeDaily, &apartment.IsFree,
			&apartment.Status, &apartment.Description, &apartment.ListingType,
			&apartment.IsAgreementAccepted, &apartment.ViewCount, &apartment.BookingCount,
			&apartment.CreatedAt, &apartment.UpdatedAt,
			&cityID, &cityName, &cityRegionID, &cityCreatedAt, &cityUpdatedAt,
			&districtID, &districtName, &districtCityID, &districtCreatedAt, &districtUpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment", "scan")
		}

		if cityID.Valid {
			city.ID = int(cityID.Int32)
			city.Name = cityName.String
			city.RegionID = int(cityRegionID.Int32)
			city.CreatedAt = cityCreatedAt.Time
			city.UpdatedAt = cityUpdatedAt.Time
			apartment.City = city
		}

		if districtID.Valid {
			district.ID = int(districtID.Int32)
			district.Name = districtName.String
			district.CityID = int(districtCityID.Int32)
			district.CreatedAt = districtCreatedAt.Time
			district.UpdatedAt = districtUpdatedAt.Time
			apartment.District = district
		}

		apartments = append(apartments, apartment)
	}

	if err = utils.CheckRowsError(rows, "cleaner apartments iteration"); err != nil {
		return nil, err
	}

	return apartments, nil
}
