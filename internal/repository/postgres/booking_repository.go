package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type bookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) domain.BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(booking *domain.Booking) error {
	query := `
		INSERT INTO bookings (
			renter_id, apartment_id, start_date, end_date, duration, cleaning_duration, status,
			total_price, service_fee, final_price, is_contract_accepted, 
			door_status, can_extend
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		booking.RenterID,
		booking.ApartmentID,
		booking.StartDate,
		booking.EndDate,
		booking.Duration,
		booking.CleaningDuration,
		booking.Status,
		booking.TotalPrice,
		booking.ServiceFee,
		booking.FinalPrice,
		booking.IsContractAccepted,
		booking.DoorStatus,
		booking.CanExtend,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "booking", "create")
	}

	bookingNumber := fmt.Sprintf("AD%03d", booking.ID)

	updateQuery := `UPDATE bookings SET booking_number = $1 WHERE id = $2`
	_, err = r.db.Exec(updateQuery, bookingNumber, booking.ID)
	if err != nil {
		return utils.HandleSQLError(err, "booking number", "update")
	}

	booking.BookingNumber = bookingNumber

	return nil
}

func (r *bookingRepository) GetByID(id int) (*domain.Booking, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		WHERE b.id = $1`, utils.BookingSelectFields)

	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "booking", "get", id)
	}
	defer utils.CloseRows(rows)

	if !rows.Next() {
		return nil, fmt.Errorf("booking with id %d not found", id)
	}

	booking, err := utils.ScanBooking(rows)
	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "booking", "get", id)
	}

	return booking, nil
}

func (r *bookingRepository) GetByBookingNumber(bookingNumber string) (*domain.Booking, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		WHERE b.booking_number = $1`, utils.BookingSelectFields)

	rows, err := r.db.Query(query, bookingNumber)
	if err != nil {
		return nil, err
	}
	defer utils.CloseRows(rows)

	if !rows.Next() {
		return nil, fmt.Errorf("бронирование с номером %s не найдено", bookingNumber)
	}

	booking, err := utils.ScanBooking(rows)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (r *bookingRepository) GetByRenterID(renterID int, status []domain.BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*domain.Booking, int, error) {
	baseArgs := []interface{}{renterID}
	statusFilter, args := utils.BuildBookingStatusFilter(status, baseArgs)

	dateFilter := ""
	if dateFrom != nil {
		dateFilter += fmt.Sprintf(" AND b.end_date >= $%d", len(args)+1)
		args = append(args, *dateFrom)
	}
	if dateTo != nil {
		dateFilter += fmt.Sprintf(" AND b.start_date <= $%d", len(args)+1)
		args = append(args, *dateTo)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM bookings b
		WHERE b.renter_id = $1%s%s`, statusFilter, dateFilter)

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "bookings count by renter", "get")
	}

	offset := (page - 1) * pageSize
	paginationArgs := append(args, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		WHERE b.renter_id = $1%s%s
		ORDER BY b.created_at DESC
		LIMIT $%d OFFSET $%d`, utils.BookingSelectFields, statusFilter, dateFilter, len(args)+1, len(args)+2)

	rows, err := r.db.Query(query, paginationArgs...)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "bookings by renter", "get")
	}
	defer utils.CloseRows(rows)

	bookings, err := utils.ScanBookings(rows)
	if err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (r *bookingRepository) GetByApartmentID(apartmentID int, status []domain.BookingStatus) ([]*domain.Booking, error) {
	baseArgs := []interface{}{apartmentID}
	statusFilter, args := utils.BuildBookingStatusFilter(status, baseArgs)

	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		WHERE b.apartment_id = $1%s
		ORDER BY b.created_at DESC`, utils.BookingSelectFields, statusFilter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "bookings by apartment", "get")
	}
	defer utils.CloseRows(rows)

	return utils.ScanBookings(rows)
}

func (r *bookingRepository) GetByOwnerID(ownerID int, status []domain.BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*domain.Booking, int, error) {
	baseArgs := []interface{}{ownerID}
	statusFilter, args := utils.BuildBookingStatusFilter(status, baseArgs)

	dateFilter := ""
	if dateFrom != nil {
		dateFilter += fmt.Sprintf(" AND b.end_date >= $%d", len(args)+1)
		args = append(args, *dateFrom)
	}
	if dateTo != nil {
		dateFilter += fmt.Sprintf(" AND b.start_date <= $%d", len(args)+1)
		args = append(args, *dateTo)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM bookings b
		JOIN apartments a ON b.apartment_id = a.id
		WHERE a.owner_id = $1%s%s`, statusFilter, dateFilter)

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "bookings count by owner", "get")
	}

	offset := (page - 1) * pageSize
	paginationArgs := append(args, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		JOIN apartments a ON b.apartment_id = a.id
		WHERE a.owner_id = $1%s%s
		ORDER BY b.created_at DESC
		LIMIT $%d OFFSET $%d`, utils.BookingSelectFields, statusFilter, dateFilter, len(args)+1, len(args)+2)

	rows, err := r.db.Query(query, paginationArgs...)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "bookings by owner", "get")
	}
	defer utils.CloseRows(rows)

	bookings, err := utils.ScanBookings(rows)
	if err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (r *bookingRepository) Update(booking *domain.Booking) error {
	query := `
		UPDATE bookings SET 
			end_date = $2, duration = $3, status = $4, cleaning_duration = $5, total_price = $6, service_fee = $7, final_price = $8,
			is_contract_accepted = $9, cancellation_reason = $10, 
			owner_comment = $11, door_status = $12, last_door_action = $13, 
			can_extend = $14, extension_requested = $15, extension_end_date = $16, 
			extension_duration = $17, extension_price = $18, payment_id = $19, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.Exec(
		query,
		booking.ID,
		booking.EndDate,
		booking.Duration,
		booking.Status,
		booking.CleaningDuration,
		booking.TotalPrice,
		booking.ServiceFee,
		booking.FinalPrice,
		booking.IsContractAccepted,
		utils.StringToSQLNullString(booking.CancellationReason),
		utils.StringToSQLNullString(booking.OwnerComment),
		booking.DoorStatus,
		utils.TimeToSQLNullTime(booking.LastDoorAction),
		booking.CanExtend,
		booking.ExtensionRequested,
		utils.TimeToSQLNullTime(booking.ExtensionEndDate),
		booking.ExtensionDuration,
		booking.ExtensionPrice,
		utils.Int64ToSQLNullInt64(booking.PaymentID),
	)

	if err != nil {
		return utils.HandleSQLErrorWithID(err, "booking", "update", booking.ID)
	}

	return nil
}

func (r *bookingRepository) UpdateDoorStatus(bookingID int, doorStatus domain.DoorStatus, lastAction *time.Time) error {
	query := `UPDATE bookings SET door_status = $1, last_door_action = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`

	_, err := r.db.Exec(query, doorStatus, utils.TimeToSQLNullTime(lastAction), bookingID)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "booking door status", "update", bookingID)
	}

	return nil
}

func (r *bookingRepository) Delete(id int) error {
	query := `DELETE FROM bookings WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "booking", "delete", id)
	}
	return nil
}

func (r *bookingRepository) CheckApartmentAvailability(apartmentID int, startDate, endDate time.Time, excludeBookingID *int) (bool, error) {
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
			(start_date >= $2 AND end_date <= $3) OR
			-- ВАЖНО: Проверяем что конец нового бронирования + время уборки не пересекается с началом существующего
			($3 + INTERVAL '60 minutes' > start_date AND $2 < start_date)
		)`

	args := []interface{}{apartmentID, startDate, endDate}

	if excludeBookingID != nil {
		query += " AND id != $4"
		args = append(args, *excludeBookingID)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, utils.HandleSQLError(err, "apartment availability", "check")
	}

	return count == 0, nil
}

func (r *bookingRepository) GetNextBookingAfterDate(apartmentID int, afterDate time.Time, excludeBookingID *int) (*domain.Booking, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		WHERE b.apartment_id = $1 
		AND b.status IN ('pending', 'approved', 'active')
		AND b.start_date > $2`, utils.BookingSelectFields)

	args := []interface{}{apartmentID, afterDate}

	if excludeBookingID != nil {
		query += " AND b.id != $3"
		args = append(args, *excludeBookingID)
	}

	query += " ORDER BY b.start_date ASC LIMIT 1"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "next booking after date", "get")
	}
	defer utils.CloseRows(rows)

	if !rows.Next() {
		return nil, nil
	}

	booking, err := utils.ScanBooking(rows)
	if err != nil {
		return nil, utils.HandleSQLError(err, "next booking after date", "scan")
	}

	return booking, nil
}

func (r *bookingRepository) CreateExtension(extension *domain.BookingExtension) error {
	query := `
		INSERT INTO booking_extensions (booking_id, duration, price, status, requested_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		extension.BookingID,
		extension.Duration,
		extension.Price,
		extension.Status,
	).Scan(&extension.ID, &extension.CreatedAt, &extension.UpdatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "booking extension", "create")
	}

	return nil
}

func (r *bookingRepository) GetExtensionsByBookingID(bookingID int) ([]*domain.BookingExtension, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM booking_extensions
		WHERE booking_id = $1
		ORDER BY created_at DESC`, utils.BookingExtensionSelectFields)

	rows, err := r.db.Query(query, bookingID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "booking extensions", "get")
	}
	defer utils.CloseRows(rows)

	return utils.ScanBookingExtensions(rows)
}

func (r *bookingRepository) GetExtensionByID(extensionID int) (*domain.BookingExtension, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM booking_extensions
		WHERE id = $1`, utils.BookingExtensionSelectFields)

	extension, err := utils.ScanBookingExtension(r.db.QueryRow(query, extensionID))
	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "booking extension", "get", extensionID)
	}

	return extension, nil
}

func (r *bookingRepository) UpdateExtension(extension *domain.BookingExtension) error {
	query := `
		UPDATE booking_extensions SET 
			status = $2, approved_at = $3, payment_id = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.Exec(query, extension.ID, extension.Status, extension.ApprovedAt, extension.PaymentID)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "booking extension", "update", extension.ID)
	}

	return nil
}

func (r *bookingRepository) CreateDoorAction(action *domain.DoorAction) error {
	query := `
		INSERT INTO door_actions (booking_id, user_id, action, success, error)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRow(
		query,
		action.BookingID,
		action.UserID,
		action.Action,
		action.Success,
		action.Error,
	).Scan(&action.ID, &action.CreatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "door action", "create")
	}

	return nil
}

func (r *bookingRepository) GetDoorActionsByBookingID(bookingID int) ([]*domain.DoorAction, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM door_actions
		WHERE booking_id = $1
		ORDER BY created_at DESC`, utils.DoorActionSelectFields)

	rows, err := r.db.Query(query, bookingID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "door actions", "get")
	}
	defer utils.CloseRows(rows)

	return utils.ScanDoorActions(rows)
}

func (r *bookingRepository) GetLastDoorAction(bookingID int) (*domain.DoorAction, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM door_actions
		WHERE booking_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, utils.DoorActionSelectFields)

	action, err := utils.ScanDoorAction(r.db.QueryRow(query, bookingID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "last door action", "get")
	}

	return action, nil
}

func (r *bookingRepository) GetByStatus(status []domain.BookingStatus) ([]*domain.Booking, error) {
	if len(status) == 0 {
		return []*domain.Booking{}, nil
	}

	statusFilter, args := utils.BuildBookingStatusFilter(status, nil)

	statusFilter = strings.TrimPrefix(statusFilter, " AND ")

	if statusFilter == "" {
		return []*domain.Booking{}, nil
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		WHERE %s
		ORDER BY b.created_at DESC`, utils.BookingSelectFields, statusFilter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleSQLError(err, "bookings by status", "get")
	}
	defer utils.CloseRows(rows)

	return utils.ScanBookings(rows)
}

func (r *bookingRepository) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.Booking, int, error) {
	whereClause := ""
	args := []interface{}{}
	argCount := 0

	if status, ok := filters["status"].(string); ok && status != "" {
		whereClause += fmt.Sprintf("b.status = $%d", argCount+1)
		args = append(args, status)
		argCount++
	}

	if statusArray, ok := filters["status"].([]domain.BookingStatus); ok && len(statusArray) > 0 {
		statusFilter, statusArgs := utils.BuildBookingStatusFilter(statusArray, nil)
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += statusFilter
		args = append(args, statusArgs...)
		argCount += len(statusArgs)
	}

	if apartmentID, ok := filters["apartment_id"].(int); ok && apartmentID > 0 {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("b.apartment_id = $%d", argCount+1)
		args = append(args, apartmentID)
		argCount++
	}

	if renterID, ok := filters["renter_id"].(int); ok && renterID > 0 {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("b.renter_id = $%d", argCount+1)
		args = append(args, renterID)
		argCount++
	}

	if renterUserID, ok := filters["renter_user_id"].(int); ok && renterUserID > 0 {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("b.renter_id IN (SELECT id FROM renters WHERE user_id = $%d)", argCount+1)
		args = append(args, renterUserID)
		argCount++
	}

	if ownerUserID, ok := filters["owner_user_id"].(int); ok && ownerUserID > 0 {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("b.apartment_id IN (SELECT a.id FROM apartments a JOIN property_owners po ON a.owner_id = po.id WHERE po.user_id = $%d)", argCount+1)
		args = append(args, ownerUserID)
		argCount++
	}

	if dateFromStr, ok := filters["date_from"].(string); ok && dateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += fmt.Sprintf("b.end_date >= $%d", argCount+1)
			args = append(args, dateFrom)
			argCount++
		}
	}

	if dateFrom, ok := filters["date_from"].(time.Time); ok {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("b.end_date >= $%d", argCount+1)
		args = append(args, dateFrom)
		argCount++
	}

	if dateToStr, ok := filters["date_to"].(string); ok && dateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += fmt.Sprintf("b.start_date <= $%d", argCount+1)
			args = append(args, dateTo)
			argCount++
		}
	}

	if dateTo, ok := filters["date_to"].(time.Time); ok {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("b.start_date <= $%d", argCount+1)
		args = append(args, dateTo)
		argCount++
	}

	if whereClause != "" {
		whereClause = "WHERE " + whereClause
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bookings b %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "bookings count", "get")
	}

	offset := (page - 1) * pageSize
	limitClause := fmt.Sprintf("LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		%s
		ORDER BY b.created_at DESC
		%s`, utils.BookingSelectFields, whereClause, limitClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "bookings", "get")
	}
	defer utils.CloseRows(rows)

	bookings, err := utils.ScanBookings(rows)
	if err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (r *bookingRepository) GetStatusStatistics() (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM bookings
		GROUP BY status`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking status statistics: %w", err)
	}
	defer rows.Close()

	statistics := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan booking status statistics: %w", err)
		}
		statistics[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return statistics, nil
}

func (r *bookingRepository) GetBookingsForDateRange(apartmentID int, startDate, endDate time.Time) ([]*domain.Booking, error) {
	query := `
		SELECT id, apartment_id, renter_id, start_date, end_date, status, 
		       cleaning_duration, created_at, updated_at
		FROM bookings 
		WHERE apartment_id = $1 
		AND status IN ('pending', 'approved', 'active')
		AND start_date < $3 
		AND end_date > $2
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(query, apartmentID, startDate, endDate)
	if err != nil {
		return nil, utils.HandleSQLError(err, "bookings", "query")
	}
	defer utils.CloseRows(rows)

	var bookings []*domain.Booking

	for rows.Next() {
		booking := &domain.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.ApartmentID,
			&booking.RenterID,
			&booking.StartDate,
			&booking.EndDate,
			&booking.Status,
			&booking.CleaningDuration,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "bookings", "scan")
		}

		bookings = append(bookings, booking)
	}

	if err = rows.Err(); err != nil {
		return nil, utils.HandleSQLError(err, "bookings", "iterate")
	}

	return bookings, nil
}

func (r *bookingRepository) CleanupExpiredBookings(batchSize int) (int, error) {
	query := `
		WITH expired_bookings AS (
			SELECT id 
			FROM bookings 
			WHERE (
				-- Обычные статусы - через день после start_date
				(status IN ('created', 'pending') AND start_date < NOW() - INTERVAL '1 day')
				OR
				-- awaiting_payment - более агрессивная очистка
				(status = 'awaiting_payment' AND (
					-- Через 30 минут после создания
					created_at < NOW() - INTERVAL '30 minutes'
					OR
					-- Через 2 часа после окончания брони
					end_date < NOW() - INTERVAL '2 hours'
				))
			)
			LIMIT $1
		)
		DELETE FROM bookings 
		WHERE id IN (SELECT id FROM expired_bookings)
	`

	result, err := r.db.Exec(query, batchSize)
	if err != nil {
		return 0, utils.HandleSQLError(err, "expired bookings", "cleanup")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, utils.HandleSQLError(err, "expired bookings cleanup", "get affected rows")
	}

	return int(rowsAffected), nil
}

func (r *bookingRepository) CleanupExpiredExtensions(batchSize int) (int, error) {
	query := `
		WITH expired_extensions AS (
			SELECT id 
			FROM booking_extensions 
			WHERE status = 'awaiting_payment'
			AND created_at < NOW() - INTERVAL '2 hours'
			LIMIT $1
		)
		DELETE FROM booking_extensions 
		WHERE id IN (SELECT id FROM expired_extensions)
	`

	result, err := r.db.Exec(query, batchSize)
	if err != nil {
		return 0, utils.HandleSQLError(err, "expired extensions", "cleanup")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, utils.HandleSQLError(err, "expired extensions cleanup", "get affected rows")
	}

	return int(rowsAffected), nil
}
