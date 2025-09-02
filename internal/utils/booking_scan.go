package utils

import (
	"database/sql"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
)

const BookingSelectFields = `
	b.id, b.renter_id, b.apartment_id, b.start_date, b.end_date, b.duration, b.cleaning_duration,
	b.status, b.total_price, b.service_fee, b.final_price, b.is_contract_accepted,
	b.cancellation_reason, b.owner_comment, b.booking_number, b.door_status,
	b.last_door_action, b.can_extend, b.extension_requested, b.extension_end_date,
	b.extension_duration, b.extension_price, b.payment_id, b.created_at, b.updated_at`

func ScanBooking(rows *sql.Rows) (*domain.Booking, error) {
	var booking domain.Booking
	var cancellationReason sql.NullString
	var ownerComment sql.NullString
	var lastDoorAction sql.NullTime
	var extensionEndDate sql.NullTime
	var paymentID sql.NullInt64

	err := rows.Scan(
		&booking.ID,
		&booking.RenterID,
		&booking.ApartmentID,
		&booking.StartDate,
		&booking.EndDate,
		&booking.Duration,
		&booking.CleaningDuration,
		&booking.Status,
		&booking.TotalPrice,
		&booking.ServiceFee,
		&booking.FinalPrice,
		&booking.IsContractAccepted,
		&cancellationReason,
		&ownerComment,
		&booking.BookingNumber,
		&booking.DoorStatus,
		&lastDoorAction,
		&booking.CanExtend,
		&booking.ExtensionRequested,
		&extensionEndDate,
		&booking.ExtensionDuration,
		&booking.ExtensionPrice,
		&paymentID,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if cancellationReason.Valid {
		booking.CancellationReason = &cancellationReason.String
	}

	if ownerComment.Valid {
		booking.OwnerComment = &ownerComment.String
	}

	if lastDoorAction.Valid {
		booking.LastDoorAction = &lastDoorAction.Time
	}

	if extensionEndDate.Valid {
		booking.ExtensionEndDate = &extensionEndDate.Time
	}

	if paymentID.Valid {
		booking.PaymentID = &paymentID.Int64
	}

	return &booking, nil
}

func ScanBookings(rows *sql.Rows) ([]*domain.Booking, error) {
	var bookings []*domain.Booking

	for rows.Next() {
		booking, err := ScanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}

	if err := CheckRowsError(rows, "booking scanning"); err != nil {
		return nil, err
	}

	if bookings == nil {
		bookings = []*domain.Booking{}
	}

	return bookings, nil
}

func BuildBookingStatusFilter(status []domain.BookingStatus, baseArgs []interface{}) (string, []interface{}) {
	if len(status) == 0 {
		return "", baseArgs
	}

	statusStrings := make([]string, len(status))
	for i, s := range status {
		statusStrings[i] = string(s)
	}

	whereClause, args := BuildInClause("b.status", statusStrings, len(baseArgs)+1)
	allArgs := append(baseArgs, args...)

	return " AND " + whereClause, allArgs
}

func BuildBookingQuery(baseWhere string, status []domain.BookingStatus, orderBy string) string {
	query := fmt.Sprintf(`
		SELECT %s
		FROM bookings b
		WHERE %s`, BookingSelectFields, baseWhere)

	if len(status) > 0 {
		statusFilter, _ := BuildBookingStatusFilter(status, nil)
		query += statusFilter
	}

	if orderBy == "" {
		orderBy = "ORDER BY b.created_at DESC"
	}

	query += " " + orderBy

	return query
}

const BookingExtensionSelectFields = `
	id, booking_id, duration, price, status, payment_id, requested_at, approved_at, created_at, updated_at`

func ScanBookingExtension(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.BookingExtension, error) {
	extension := &domain.BookingExtension{}

	err := scanner.Scan(
		&extension.ID,
		&extension.BookingID,
		&extension.Duration,
		&extension.Price,
		&extension.Status,
		&extension.PaymentID,
		&extension.RequestedAt,
		&extension.ApprovedAt,
		&extension.CreatedAt,
		&extension.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return extension, nil
}

func ScanBookingExtensions(rows *sql.Rows) ([]*domain.BookingExtension, error) {
	var extensions []*domain.BookingExtension

	for rows.Next() {
		extension, err := ScanBookingExtension(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking extension: %w", err)
		}
		extensions = append(extensions, extension)
	}

	if err := CheckRowsError(rows, "booking extension scanning"); err != nil {
		return nil, err
	}

	return extensions, nil
}

const DoorActionSelectFields = `
	id, booking_id, user_id, action, success, error, created_at`

func ScanDoorAction(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.DoorAction, error) {
	action := &domain.DoorAction{}

	err := scanner.Scan(
		&action.ID,
		&action.BookingID,
		&action.UserID,
		&action.Action,
		&action.Success,
		&action.Error,
		&action.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return action, nil
}

func ScanDoorActions(rows *sql.Rows) ([]*domain.DoorAction, error) {
	var actions []*domain.DoorAction

	for rows.Next() {
		action, err := ScanDoorAction(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan door action: %w", err)
		}
		actions = append(actions, action)
	}

	if err := CheckRowsError(rows, "door action scanning"); err != nil {
		return nil, err
	}

	return actions, nil
}
