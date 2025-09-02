package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/russo2642/renti_kz/internal/domain"
)

type chatMessageRepository struct {
	db *sql.DB
}

func NewChatMessageRepository(db *sql.DB) domain.ChatMessageRepository {
	return &chatMessageRepository{
		db: db,
	}
}

func (r *chatMessageRepository) Create(message *domain.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (chat_room_id, sender_id, type, content, file_url, file_name, file_size, status, reply_to_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		message.ChatRoomID,
		message.SenderID,
		message.Type,
		message.Content,
		message.FileURL,
		message.FileName,
		message.FileSize,
		message.Status,
		message.ReplyToID,
	).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create chat message: %w", err)
	}

	return nil
}

func (r *chatMessageRepository) GetByID(id int) (*domain.ChatMessage, error) {
	query := `
		SELECT cm.id, cm.chat_room_id, cm.sender_id, cm.type, cm.content, cm.file_url, cm.file_name, cm.file_size,
			   cm.status, cm.read_at, cm.reply_to_id, cm.edited_at, cm.created_at, cm.updated_at,
			   u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.created_at, u.updated_at,
			   COALESCE(r.name, '') as role_name,
			   COALESCE(reply.id, 0), COALESCE(reply.content, ''), COALESCE(reply.type, ''), COALESCE(reply.created_at, '1970-01-01'::timestamp)
		FROM chat_messages cm
		LEFT JOIN users u ON cm.sender_id = u.id
		LEFT JOIN user_roles r ON u.role_id = r.id
		LEFT JOIN chat_messages reply ON cm.reply_to_id = reply.id
		WHERE cm.id = $1`

	message := &domain.ChatMessage{}
	sender := &domain.User{}
	var replyID int
	var replyContent, replyType string
	var replyCreatedAt sql.NullTime

	var roleName string

	err := r.db.QueryRow(query, id).Scan(
		&message.ID, &message.ChatRoomID, &message.SenderID, &message.Type, &message.Content,
		&message.FileURL, &message.FileName, &message.FileSize, &message.Status, &message.ReadAt,
		&message.ReplyToID, &message.EditedAt, &message.CreatedAt, &message.UpdatedAt,
		&sender.ID, &sender.Phone, &sender.FirstName, &sender.LastName, &sender.Email,
		&sender.CityID, &sender.IIN, &sender.RoleID, &sender.CreatedAt, &sender.UpdatedAt,
		&roleName,
		&replyID, &replyContent, &replyType, &replyCreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat message not found")
		}
		return nil, fmt.Errorf("failed to get chat message: %w", err)
	}

	sender.Role = domain.UserRole(roleName)
	message.Sender = sender

	if replyID > 0 {
		replyMessage := &domain.ChatMessage{
			ID:      replyID,
			Content: replyContent,
			Type:    domain.MessageType(replyType),
		}
		if replyCreatedAt.Valid {
			replyMessage.CreatedAt = replyCreatedAt.Time
		}
		message.ReplyTo = replyMessage
	}

	return message, nil
}

func (r *chatMessageRepository) GetByRoomID(roomID int, page, pageSize int) ([]*domain.ChatMessage, int, error) {
	countQuery := `SELECT COUNT(*) FROM chat_messages WHERE chat_room_id = $1`
	var total int
	err := r.db.QueryRow(countQuery, roomID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count chat messages: %w", err)
	}

	query := `
		SELECT cm.id, cm.chat_room_id, cm.sender_id, cm.type, cm.content, cm.file_url, cm.file_name, cm.file_size,
			   cm.status, cm.read_at, cm.reply_to_id, cm.edited_at, cm.created_at, cm.updated_at,
			   u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.created_at, u.updated_at,
			   COALESCE(r.name, '') as role_name
		FROM chat_messages cm
		LEFT JOIN users u ON cm.sender_id = u.id
		LEFT JOIN user_roles r ON u.role_id = r.id
		WHERE cm.chat_room_id = $1
		ORDER BY cm.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, roomID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chat messages by room ID: %w", err)
	}
	defer rows.Close()

	var messages []*domain.ChatMessage = make([]*domain.ChatMessage, 0)

	for rows.Next() {
		message := &domain.ChatMessage{}
		sender := &domain.User{}
		var roleName string

		err := rows.Scan(
			&message.ID, &message.ChatRoomID, &message.SenderID, &message.Type, &message.Content,
			&message.FileURL, &message.FileName, &message.FileSize, &message.Status, &message.ReadAt,
			&message.ReplyToID, &message.EditedAt, &message.CreatedAt, &message.UpdatedAt,
			&sender.ID, &sender.Phone, &sender.FirstName, &sender.LastName, &sender.Email,
			&sender.CityID, &sender.IIN, &sender.RoleID, &sender.CreatedAt, &sender.UpdatedAt,
			&roleName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat message: %w", err)
		}

		sender.Role = domain.UserRole(roleName)
		message.Sender = sender
		messages = append(messages, message)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, nil
}

func (r *chatMessageRepository) GetUnreadCount(roomID, userID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM chat_messages cm
		WHERE cm.chat_room_id = $1 
		  AND cm.sender_id != $2 
		  AND cm.status != 'read'
		  AND NOT EXISTS (
		      SELECT 1 FROM chat_participants cp 
		      WHERE cp.chat_room_id = $1 
		        AND cp.user_id = $2 
		        AND cp.last_read_at >= cm.created_at
		  )`

	var count int
	err := r.db.QueryRow(query, roomID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread messages count: %w", err)
	}

	return count, nil
}

func (r *chatMessageRepository) GetUnreadMessages(roomID, userID int) ([]*domain.ChatMessage, error) {
	query := `
		SELECT cm.id, cm.chat_room_id, cm.sender_id, cm.type, cm.content, cm.file_url, cm.file_name, cm.file_size,
			   cm.status, cm.read_at, cm.reply_to_id, cm.edited_at, cm.created_at, cm.updated_at,
			   u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.created_at, u.updated_at,
			   COALESCE(r.name, '') as role_name
		FROM chat_messages cm
		LEFT JOIN users u ON cm.sender_id = u.id
		LEFT JOIN user_roles r ON u.role_id = r.id
		WHERE cm.chat_room_id = $1 
		  AND cm.sender_id != $2 
		  AND cm.status != 'read'
		  AND NOT EXISTS (
		      SELECT 1 FROM chat_participants cp 
		      WHERE cp.chat_room_id = $1 
		        AND cp.user_id = $2 
		        AND cp.last_read_at >= cm.created_at
		  )
		ORDER BY cm.created_at ASC`

	rows, err := r.db.Query(query, roomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.ChatMessage = make([]*domain.ChatMessage, 0)

	for rows.Next() {
		message := &domain.ChatMessage{}
		sender := &domain.User{}
		var roleName string

		err := rows.Scan(
			&message.ID, &message.ChatRoomID, &message.SenderID, &message.Type, &message.Content,
			&message.FileURL, &message.FileName, &message.FileSize, &message.Status, &message.ReadAt,
			&message.ReplyToID, &message.EditedAt, &message.CreatedAt, &message.UpdatedAt,
			&sender.ID, &sender.Phone, &sender.FirstName, &sender.LastName, &sender.Email,
			&sender.CityID, &sender.IIN, &sender.RoleID, &sender.CreatedAt, &sender.UpdatedAt,
			&roleName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unread message: %w", err)
		}

		sender.Role = domain.UserRole(roleName)
		message.Sender = sender
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *chatMessageRepository) Update(message *domain.ChatMessage) error {
	query := `
		UPDATE chat_messages 
		SET content = $2, edited_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING edited_at, updated_at`

	err := r.db.QueryRow(
		query,
		message.ID,
		message.Content,
	).Scan(&message.EditedAt, &message.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update chat message: %w", err)
	}

	return nil
}

func (r *chatMessageRepository) Delete(id int) error {
	query := `DELETE FROM chat_messages WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat message not found")
	}

	return nil
}

func (r *chatMessageRepository) MarkAsRead(messageIDs []int, userID int) error {
	if len(messageIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(messageIDs))
	args := make([]interface{}, len(messageIDs)+1)
	args[0] = userID

	for i, id := range messageIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		UPDATE chat_messages 
		SET status = 'read', read_at = NOW(), updated_at = NOW()
		WHERE id IN (%s) AND sender_id != $1`,
		strings.Join(placeholders, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

func (r *chatMessageRepository) MarkRoomAsRead(roomID, userID int) error {
	query := `
		UPDATE chat_participants 
		SET last_read_at = NOW() 
		WHERE chat_room_id = $1 AND user_id = $2`

	_, err := r.db.Exec(query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark room as read: %w", err)
	}

	return nil
}

func (r *chatMessageRepository) GetLastMessage(roomID int) (*domain.ChatMessage, error) {
	query := `
		SELECT cm.id, cm.chat_room_id, cm.sender_id, cm.type, cm.content, cm.file_url, cm.file_name, cm.file_size,
			   cm.status, cm.read_at, cm.reply_to_id, cm.edited_at, cm.created_at, cm.updated_at,
			   u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, u.role_id, u.created_at, u.updated_at,
			   COALESCE(r.name, '') as role_name
		FROM chat_messages cm
		LEFT JOIN users u ON cm.sender_id = u.id
		LEFT JOIN user_roles r ON u.role_id = r.id
		WHERE cm.chat_room_id = $1
		ORDER BY cm.created_at DESC
		LIMIT 1`

	message := &domain.ChatMessage{}
	sender := &domain.User{}
	var roleName string

	err := r.db.QueryRow(query, roomID).Scan(
		&message.ID, &message.ChatRoomID, &message.SenderID, &message.Type, &message.Content,
		&message.FileURL, &message.FileName, &message.FileSize, &message.Status, &message.ReadAt,
		&message.ReplyToID, &message.EditedAt, &message.CreatedAt, &message.UpdatedAt,
		&sender.ID, &sender.Phone, &sender.FirstName, &sender.LastName, &sender.Email,
		&sender.CityID, &sender.IIN, &sender.RoleID, &sender.CreatedAt, &sender.UpdatedAt,
		&roleName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}

	sender.Role = domain.UserRole(roleName)
	message.Sender = sender

	return message, nil
}
