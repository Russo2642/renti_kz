package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/russo2642/renti_kz/internal/domain"
)

type chatRoomRepository struct {
	db *sql.DB
}

func NewChatRoomRepository(db *sql.DB) domain.ChatRoomRepository {
	return &chatRoomRepository{
		db: db,
	}
}

func (r *chatRoomRepository) Create(room *domain.ChatRoom) error {
	query := `
		INSERT INTO chat_rooms (booking_id, concierge_id, renter_id, apartment_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		room.BookingID,
		room.ConciergeID,
		room.RenterID,
		room.ApartmentID,
		room.Status,
	).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create chat room: %w", err)
	}

	return nil
}

func (r *chatRoomRepository) GetByID(id int) (*domain.ChatRoom, error) {
	query := `
		SELECT cr.id, cr.booking_id, cr.concierge_id, cr.renter_id, cr.apartment_id, 
			   cr.status, cr.opened_at, cr.closed_at, cr.last_message_at, cr.created_at, cr.updated_at,
			   b.id, b.start_date, b.end_date, b.total_price, b.status, b.created_at, b.updated_at,
			   c.id, c.user_id, c.is_active, c.created_at, c.updated_at,
			   cu.id, cu.first_name, cu.last_name, cu.email, cu.phone,
			   r.id, r.user_id, r.created_at, r.updated_at,
			   ru.id, ru.first_name, ru.last_name, ru.email, ru.phone,
			   a.id, a.description, a.street, a.building, a.city_id, a.price, a.is_free
		FROM chat_rooms cr
		LEFT JOIN bookings b ON cr.booking_id = b.id
		LEFT JOIN concierges c ON cr.concierge_id = c.id
		LEFT JOIN users cu ON c.user_id = cu.id
		LEFT JOIN renters r ON cr.renter_id = r.id
		LEFT JOIN users ru ON r.user_id = ru.id
		LEFT JOIN apartments a ON cr.apartment_id = a.id
		WHERE cr.id = $1`

	room := &domain.ChatRoom{}
	booking := &domain.Booking{}
	concierge := &domain.Concierge{}
	conciergeUser := &domain.User{}
	renter := &domain.Renter{}
	renterUser := &domain.User{}
	apartment := &domain.Apartment{}

	err := r.db.QueryRow(query, id).Scan(
		&room.ID, &room.BookingID, &room.ConciergeID, &room.RenterID, &room.ApartmentID,
		&room.Status, &room.OpenedAt, &room.ClosedAt, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt,
		&booking.ID, &booking.StartDate, &booking.EndDate, &booking.TotalPrice, &booking.Status,
		&booking.CreatedAt, &booking.UpdatedAt,
		&concierge.ID, &concierge.UserID, &concierge.IsActive,
		&concierge.CreatedAt, &concierge.UpdatedAt,
		&conciergeUser.ID, &conciergeUser.FirstName, &conciergeUser.LastName, &conciergeUser.Email, &conciergeUser.Phone,
		&renter.ID, &renter.UserID, &renter.CreatedAt, &renter.UpdatedAt,
		&renterUser.ID, &renterUser.FirstName, &renterUser.LastName, &renterUser.Email, &renterUser.Phone,
		&apartment.ID, &apartment.Description, &apartment.Street, &apartment.Building,
		&apartment.CityID, &apartment.Price, &apartment.IsFree,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat room not found")
		}
		return nil, fmt.Errorf("failed to get chat room: %w", err)
	}

	room.Booking = booking
	concierge.User = conciergeUser
	room.Concierge = concierge
	renter.User = renterUser
	room.Renter = renter
	room.Apartment = apartment

	return room, nil
}

func (r *chatRoomRepository) GetByBookingID(bookingID int) (*domain.ChatRoom, error) {
	query := `
		SELECT cr.id, cr.booking_id, cr.concierge_id, cr.renter_id, cr.apartment_id, 
			   cr.status, cr.opened_at, cr.closed_at, cr.last_message_at, cr.created_at, cr.updated_at
		FROM chat_rooms cr
		WHERE cr.booking_id = $1`

	room := &domain.ChatRoom{}

	err := r.db.QueryRow(query, bookingID).Scan(
		&room.ID, &room.BookingID, &room.ConciergeID, &room.RenterID, &room.ApartmentID,
		&room.Status, &room.OpenedAt, &room.ClosedAt, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat room not found")
		}
		return nil, fmt.Errorf("failed to get chat room by booking ID: %w", err)
	}

	return room, nil
}

func (r *chatRoomRepository) GetByUserID(userID int, status []domain.ChatRoomStatus, page, pageSize int) ([]*domain.ChatRoom, int, error) {
	whereConditions := []string{
		"(ru.id = $1 OR cu.id = $1)",
	}
	args := []interface{}{userID}
	argIndex := 2

	if len(status) > 0 {
		statusPlaceholders := make([]string, len(status))
		for i, s := range status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, string(s))
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("cr.status IN (%s)", strings.Join(statusPlaceholders, ", ")))
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	baseQuery := `
		FROM chat_rooms cr
		LEFT JOIN renters r ON cr.renter_id = r.id
		LEFT JOIN users ru ON r.user_id = ru.id
		LEFT JOIN concierges c ON cr.concierge_id = c.id
		LEFT JOIN users cu ON c.user_id = cu.id`

	countQuery := "SELECT COUNT(*) " + baseQuery + " " + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count chat rooms: %w", err)
	}

	query := `
		SELECT cr.id, cr.booking_id, cr.concierge_id, cr.renter_id, cr.apartment_id, 
			   cr.status, cr.opened_at, cr.closed_at, cr.last_message_at, cr.created_at, cr.updated_at
		` + baseQuery + " " + whereClause + `
		ORDER BY cr.last_message_at DESC NULLS LAST, cr.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chat rooms by user ID: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.ChatRoom = make([]*domain.ChatRoom, 0)

	for rows.Next() {
		room := &domain.ChatRoom{}

		err := rows.Scan(
			&room.ID, &room.BookingID, &room.ConciergeID, &room.RenterID, &room.ApartmentID,
			&room.Status, &room.OpenedAt, &room.ClosedAt, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat room: %w", err)
		}

		rooms = append(rooms, room)
	}

	return rooms, total, nil
}

func (r *chatRoomRepository) GetByConciergeID(conciergeID int, status []domain.ChatRoomStatus, page, pageSize int) ([]*domain.ChatRoom, int, error) {
	whereConditions := []string{"cr.concierge_id = $1"}
	args := []interface{}{conciergeID}
	argIndex := 2

	if len(status) > 0 {
		statusPlaceholders := make([]string, len(status))
		for i, s := range status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, string(s))
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("cr.status IN (%s)", strings.Join(statusPlaceholders, ", ")))
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	baseQuery := "FROM chat_rooms cr"
	countQuery := "SELECT COUNT(*) " + baseQuery + " " + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count chat rooms: %w", err)
	}

	query := `
		SELECT cr.id, cr.booking_id, cr.concierge_id, cr.renter_id, cr.apartment_id, 
			   cr.status, cr.opened_at, cr.closed_at, cr.last_message_at, cr.created_at, cr.updated_at
		` + baseQuery + " " + whereClause + `
		ORDER BY cr.last_message_at DESC NULLS LAST, cr.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chat rooms by concierge ID: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.ChatRoom = make([]*domain.ChatRoom, 0)

	for rows.Next() {
		room := &domain.ChatRoom{}

		err := rows.Scan(
			&room.ID, &room.BookingID, &room.ConciergeID, &room.RenterID, &room.ApartmentID,
			&room.Status, &room.OpenedAt, &room.ClosedAt, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat room: %w", err)
		}

		rooms = append(rooms, room)
	}

	return rooms, total, nil
}

func (r *chatRoomRepository) GetActiveRooms() ([]*domain.ChatRoom, error) {
	query := `
		SELECT id, booking_id, concierge_id, renter_id, apartment_id, 
			   status, opened_at, closed_at, last_message_at, created_at, updated_at
		FROM chat_rooms
		WHERE status = 'active'
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active chat rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.ChatRoom = make([]*domain.ChatRoom, 0)

	for rows.Next() {
		room := &domain.ChatRoom{}

		err := rows.Scan(
			&room.ID, &room.BookingID, &room.ConciergeID, &room.RenterID, &room.ApartmentID,
			&room.Status, &room.OpenedAt, &room.ClosedAt, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat room: %w", err)
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (r *chatRoomRepository) GetRoomsPendingOpen() ([]*domain.ChatRoom, error) {
	query := `
		SELECT cr.id, cr.booking_id, cr.concierge_id, cr.renter_id, cr.apartment_id, 
			   cr.status, cr.opened_at, cr.closed_at, cr.last_message_at, cr.created_at, cr.updated_at
		FROM chat_rooms cr
		JOIN bookings b ON cr.booking_id = b.id
		WHERE cr.status = 'pending' AND (b.start_date - INTERVAL '15 minutes') <= NOW()
		ORDER BY b.start_date ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms pending open: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.ChatRoom = make([]*domain.ChatRoom, 0)

	for rows.Next() {
		room := &domain.ChatRoom{}

		err := rows.Scan(
			&room.ID, &room.BookingID, &room.ConciergeID, &room.RenterID, &room.ApartmentID,
			&room.Status, &room.OpenedAt, &room.ClosedAt, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat room: %w", err)
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (r *chatRoomRepository) GetRoomsPendingClose() ([]*domain.ChatRoom, error) {
	query := `
		SELECT cr.id, cr.booking_id, cr.concierge_id, cr.renter_id, cr.apartment_id, 
			   cr.status, cr.opened_at, cr.closed_at, cr.last_message_at, cr.created_at, cr.updated_at
		FROM chat_rooms cr
		JOIN bookings b ON cr.booking_id = b.id
		WHERE cr.status = 'active' AND (b.end_date + INTERVAL '24 hours') <= NOW()
		ORDER BY b.end_date ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms pending close: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.ChatRoom = make([]*domain.ChatRoom, 0)

	for rows.Next() {
		room := &domain.ChatRoom{}

		err := rows.Scan(
			&room.ID, &room.BookingID, &room.ConciergeID, &room.RenterID, &room.ApartmentID,
			&room.Status, &room.OpenedAt, &room.ClosedAt, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat room: %w", err)
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (r *chatRoomRepository) Update(room *domain.ChatRoom) error {
	query := `
		UPDATE chat_rooms 
		SET status = $2, opened_at = $3, closed_at = $4, last_message_at = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRow(
		query,
		room.ID,
		room.Status,
		room.OpenedAt,
		room.ClosedAt,
		room.LastMessageAt,
	).Scan(&room.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update chat room: %w", err)
	}

	return nil
}

func (r *chatRoomRepository) Delete(id int) error {
	query := `DELETE FROM chat_rooms WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat room not found")
	}

	return nil
}

func (r *chatRoomRepository) CanUserAccessRoom(roomID, userID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM chat_rooms cr
			LEFT JOIN renters r ON cr.renter_id = r.id
			LEFT JOIN concierges c ON cr.concierge_id = c.id
			WHERE cr.id = $1 AND (r.user_id = $2 OR c.user_id = $2)
		)`

	var canAccess bool
	err := r.db.QueryRow(query, roomID, userID).Scan(&canAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check user access to room: %w", err)
	}

	return canAccess, nil
}
