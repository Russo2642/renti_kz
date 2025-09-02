package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type chatParticipantRepository struct {
	db *sql.DB
}

func NewChatParticipantRepository(db *sql.DB) domain.ChatParticipantRepository {
	return &chatParticipantRepository{
		db: db,
	}
}

func (r *chatParticipantRepository) Create(participant *domain.ChatParticipant) error {
	query := `
		INSERT INTO chat_participants (chat_room_id, user_id, role, joined_at, is_online, last_seen_at)
		VALUES ($1, $2, $3, NOW(), $4, NOW())
		RETURNING id, joined_at, last_seen_at`

	err := r.db.QueryRow(
		query,
		participant.ChatRoomID,
		participant.UserID,
		participant.Role,
		participant.IsOnline,
	).Scan(&participant.ID, &participant.JoinedAt, &participant.LastSeenAt)

	if err != nil {
		return fmt.Errorf("failed to create chat participant: %w", err)
	}

	return nil
}

func (r *chatParticipantRepository) GetByRoomID(roomID int) ([]*domain.ChatParticipant, error) {
	query := `
		SELECT cp.id, cp.chat_room_id, cp.user_id, cp.role, cp.joined_at, cp.left_at, cp.last_read_at, cp.is_online, cp.last_seen_at,
			   u.id, u.first_name, u.last_name, u.email, u.phone, u.created_at, u.updated_at
		FROM chat_participants cp
		LEFT JOIN users u ON cp.user_id = u.id
		WHERE cp.chat_room_id = $1 AND cp.left_at IS NULL
		ORDER BY cp.joined_at ASC`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat participants by room ID: %w", err)
	}
	defer rows.Close()

	var participants []*domain.ChatParticipant

	for rows.Next() {
		participant := &domain.ChatParticipant{}
		user := &domain.User{}

		err := rows.Scan(
			&participant.ID, &participant.ChatRoomID, &participant.UserID, &participant.Role,
			&participant.JoinedAt, &participant.LeftAt, &participant.LastReadAt, &participant.IsOnline, &participant.LastSeenAt,
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat participant: %w", err)
		}

		participant.User = user
		participants = append(participants, participant)
	}

	return participants, nil
}

func (r *chatParticipantRepository) GetByUserAndRoom(userID, roomID int) (*domain.ChatParticipant, error) {
	query := `
		SELECT cp.id, cp.chat_room_id, cp.user_id, cp.role, cp.joined_at, cp.left_at, cp.last_read_at, cp.is_online, cp.last_seen_at,
			   u.id, u.first_name, u.last_name, u.email, u.phone, u.created_at, u.updated_at
		FROM chat_participants cp
		LEFT JOIN users u ON cp.user_id = u.id
		WHERE cp.user_id = $1 AND cp.chat_room_id = $2`

	participant := &domain.ChatParticipant{}
	user := &domain.User{}

	err := r.db.QueryRow(query, userID, roomID).Scan(
		&participant.ID, &participant.ChatRoomID, &participant.UserID, &participant.Role,
		&participant.JoinedAt, &participant.LeftAt, &participant.LastReadAt, &participant.IsOnline, &participant.LastSeenAt,
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat participant not found")
		}
		return nil, fmt.Errorf("failed to get chat participant: %w", err)
	}

	participant.User = user

	return participant, nil
}

func (r *chatParticipantRepository) UpdateLastSeen(userID, roomID int, lastSeen time.Time) error {
	query := `
		UPDATE chat_participants 
		SET last_seen_at = $3
		WHERE user_id = $1 AND chat_room_id = $2`

	result, err := r.db.Exec(query, userID, roomID, lastSeen)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat participant not found")
	}

	return nil
}

func (r *chatParticipantRepository) UpdateOnlineStatus(userID, roomID int, isOnline bool) error {
	query := `
		UPDATE chat_participants 
		SET is_online = $3, last_seen_at = NOW()
		WHERE user_id = $1 AND chat_room_id = $2`

	result, err := r.db.Exec(query, userID, roomID, isOnline)
	if err != nil {
		return fmt.Errorf("failed to update online status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return r.createParticipantIfNotExists(userID, roomID, isOnline)
	}

	return nil
}

func (r *chatParticipantRepository) UpdateLastRead(userID, roomID int, lastRead time.Time) error {
	query := `
		UPDATE chat_participants 
		SET last_read_at = $3
		WHERE user_id = $1 AND chat_room_id = $2`

	result, err := r.db.Exec(query, userID, roomID, lastRead)
	if err != nil {
		return fmt.Errorf("failed to update last read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return r.createParticipantIfNotExists(userID, roomID, false)
	}

	return nil
}

func (r *chatParticipantRepository) createParticipantIfNotExists(userID, roomID int, isOnline bool) error {
	roleQuery := `
		SELECT CASE 
			WHEN r.user_id = $1 THEN 'renter'
			WHEN c.user_id = $1 THEN 'concierge'
			ELSE NULL
		END as role
		FROM chat_rooms cr
		LEFT JOIN renters r ON cr.renter_id = r.id
		LEFT JOIN concierges c ON cr.concierge_id = c.id
		WHERE cr.id = $2`

	var role sql.NullString
	err := r.db.QueryRow(roleQuery, userID, roomID).Scan(&role)
	if err != nil || !role.Valid {
		return fmt.Errorf("user is not authorized for this chat room")
	}

	insertQuery := `
		INSERT INTO chat_participants (chat_room_id, user_id, role, joined_at, is_online, last_seen_at, last_read_at)
		VALUES ($1, $2, $3, NOW(), $4, NOW(), NOW())
		ON CONFLICT (chat_room_id, user_id) DO UPDATE SET
			is_online = EXCLUDED.is_online,
			last_seen_at = EXCLUDED.last_seen_at,
			last_read_at = EXCLUDED.last_read_at`

	_, err = r.db.Exec(insertQuery, roomID, userID, role.String, isOnline)
	if err != nil {
		return fmt.Errorf("failed to create chat participant: %w", err)
	}

	return nil
}
