package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

func (r *NotificationRepository) CreateNotification(notification *domain.Notification) error {
	var dataJSON []byte
	var err error

	if notification.Data != nil {
		dataJSON, err = json.Marshal(notification.Data)
		if err != nil {
			return fmt.Errorf("ошибка сериализации data: %w", err)
		}
	}

	query := `
		INSERT INTO notifications (
			user_id, type, priority, title, message, data, 
			is_read, is_pushed, booking_id, apartment_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err = r.db.QueryRow(
		query,
		notification.UserID,
		notification.Type,
		notification.Priority,
		notification.Title,
		notification.Message,
		dataJSON,
		notification.IsRead,
		notification.IsPushed,
		notification.BookingID,
		notification.ApartmentID,
	).Scan(&notification.ID, &notification.CreatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "notification", "create")
	}

	return nil
}

func (r *NotificationRepository) GetNotificationByID(id int) (*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, priority, title, message, data,
			   is_read, is_pushed, booking_id, apartment_id, 
			   created_at, read_at
		FROM notifications 
		WHERE id = $1`

	notification, err := r.scanNotification(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLErrorWithID(err, "notification", "get", id)
	}

	return notification, nil
}

func (r *NotificationRepository) GetUserNotifications(userID int, limit, offset int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, priority, title, message, data,
			   is_read, is_pushed, booking_id, apartment_id, 
			   created_at, read_at
		FROM notifications 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, utils.HandleSQLError(err, "notifications by user", "query")
	}
	defer utils.CloseRows(rows)

	return r.scanNotifications(rows)
}

func (r *NotificationRepository) GetUnreadNotifications(userID int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, priority, title, message, data,
			   is_read, is_pushed, booking_id, apartment_id, 
			   created_at, read_at
		FROM notifications 
		WHERE user_id = $1 AND is_read = FALSE
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "unread notifications", "query")
	}
	defer utils.CloseRows(rows)

	return r.scanNotifications(rows)
}

func (r *NotificationRepository) MarkAsRead(notificationID, userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = TRUE, read_at = NOW() 
		WHERE id = $1 AND user_id = $2`

	_, err := r.db.Exec(query, notificationID, userID)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "notification", "mark as read", notificationID)
	}

	return nil
}

func (r *NotificationRepository) MarkMultipleAsRead(notificationIDs []int, userID int) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	query := `
		UPDATE notifications 
		SET is_read = TRUE, read_at = NOW() 
		WHERE id = ANY($1) AND user_id = $2`

	_, err := r.db.Exec(query, pq.Array(notificationIDs), userID)
	if err != nil {
		return utils.HandleSQLError(err, "notifications", "mark multiple as read")
	}

	return nil
}

func (r *NotificationRepository) MarkAllAsRead(userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = TRUE, read_at = NOW() 
		WHERE user_id = $1 AND is_read = FALSE`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return utils.HandleSQLError(err, "notifications", "mark all as read for user")
	}

	return nil
}

func (r *NotificationRepository) GetUnreadCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, utils.HandleSQLError(err, "unread notifications count", "query")
	}

	return count, nil
}

func (r *NotificationRepository) DeleteNotification(id int) error {
	query := `DELETE FROM notifications WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "notification", "delete", id)
	}

	return nil
}

func (r *NotificationRepository) DeleteNotificationsByUser(notificationID, userID int) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, notificationID, userID)
	if err != nil {
		return utils.HandleSQLError(err, "notification", "delete_by_user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("уведомление не найдено или не принадлежит пользователю")
	}

	return nil
}

func (r *NotificationRepository) DeleteReadNotifications(userID int) (int, error) {
	query := `DELETE FROM notifications WHERE user_id = $1 AND is_read = true`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return 0, utils.HandleSQLError(err, "notifications", "delete_read")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}

	return int(rowsAffected), nil
}

func (r *NotificationRepository) DeleteAllNotifications(userID int) (int, error) {
	query := `DELETE FROM notifications WHERE user_id = $1`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return 0, utils.HandleSQLError(err, "notifications", "delete_all")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}

	return int(rowsAffected), nil
}

func (r *NotificationRepository) RegisterDevice(device *domain.UserDevice) error {
	query := `
		INSERT INTO user_devices (
			user_id, device_token, device_type, is_active,
			app_version, os_version, last_seen_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (device_token) 
		DO UPDATE SET 
			user_id = EXCLUDED.user_id,
			device_type = EXCLUDED.device_type,
			is_active = EXCLUDED.is_active,
			app_version = EXCLUDED.app_version,
			os_version = EXCLUDED.os_version,
			last_seen_at = EXCLUDED.last_seen_at,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		device.UserID,
		device.DeviceToken,
		device.DeviceType,
		device.IsActive,
		device.AppVersion,
		device.OSVersion,
		device.LastSeenAt,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "device", "register")
	}

	return nil
}

func (r *NotificationRepository) GetDevicesByUserID(userID int) ([]*domain.UserDevice, error) {
	query := `
		SELECT id, user_id, device_token, device_type, is_active,
			   app_version, os_version, last_seen_at, created_at, updated_at
		FROM user_devices 
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY last_seen_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "devices by user", "query")
	}
	defer utils.CloseRows(rows)

	return r.scanDevices(rows)
}

func (r *NotificationRepository) GetDeviceByToken(token string) (*domain.UserDevice, error) {
	query := `
		SELECT id, user_id, device_token, device_type, is_active,
			   app_version, os_version, last_seen_at, created_at, updated_at
		FROM user_devices 
		WHERE device_token = $1`

	device, err := r.scanDevice(r.db.QueryRow(query, token))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "device by token", "get")
	}

	return device, nil
}

func (r *NotificationRepository) UpdateDevice(device *domain.UserDevice) error {
	query := `
		UPDATE user_devices SET
			user_id = $2,
			device_type = $3,
			is_active = $4,
			app_version = $5,
			os_version = $6,
			last_seen_at = $7,
			updated_at = NOW()
		WHERE device_token = $1`

	_, err := r.db.Exec(
		query,
		device.DeviceToken,
		device.UserID,
		device.DeviceType,
		device.IsActive,
		device.AppVersion,
		device.OSVersion,
		device.LastSeenAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "device", "update")
	}

	return nil
}

func (r *NotificationRepository) DeactivateDevice(token string) error {
	query := `
		UPDATE user_devices 
		SET is_active = FALSE, updated_at = NOW() 
		WHERE device_token = $1`

	_, err := r.db.Exec(query, token)
	if err != nil {
		return utils.HandleSQLError(err, "device", "deactivate")
	}

	return nil
}

func (r *NotificationRepository) UpdateDeviceHeartbeat(token string) error {
	query := `
		UPDATE user_devices 
		SET last_seen_at = NOW(), updated_at = NOW() 
		WHERE device_token = $1 AND is_active = TRUE`

	_, err := r.db.Exec(query, token)
	if err != nil {
		return utils.HandleSQLError(err, "device heartbeat", "update")
	}

	return nil
}

func (r *NotificationRepository) scanNotification(row *sql.Row) (*domain.Notification, error) {
	var notification domain.Notification
	var dataJSON []byte

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Priority,
		&notification.Title,
		&notification.Message,
		&dataJSON,
		&notification.IsRead,
		&notification.IsPushed,
		&notification.BookingID,
		&notification.ApartmentID,
		&notification.CreatedAt,
		&notification.ReadAt,
	)

	if err != nil {
		return nil, err
	}

	if len(dataJSON) > 0 {
		err = json.Unmarshal(dataJSON, &notification.Data)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации data: %w", err)
		}
	}

	return &notification, nil
}

func (r *NotificationRepository) scanNotifications(rows *sql.Rows) ([]*domain.Notification, error) {
	var notifications []*domain.Notification

	for rows.Next() {
		var notification domain.Notification
		var dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Priority,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.IsRead,
			&notification.IsPushed,
			&notification.BookingID,
			&notification.ApartmentID,
			&notification.CreatedAt,
			&notification.ReadAt,
		)

		if err != nil {
			return nil, err
		}

		if len(dataJSON) > 0 {
			err = json.Unmarshal(dataJSON, &notification.Data)
			if err != nil {
				return nil, fmt.Errorf("ошибка десериализации data: %w", err)
			}
		}

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) scanDevice(row *sql.Row) (*domain.UserDevice, error) {
	var device domain.UserDevice

	err := row.Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceToken,
		&device.DeviceType,
		&device.IsActive,
		&device.AppVersion,
		&device.OSVersion,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &device, nil
}

func (r *NotificationRepository) scanDevices(rows *sql.Rows) ([]*domain.UserDevice, error) {
	var devices []*domain.UserDevice

	for rows.Next() {
		var device domain.UserDevice

		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceToken,
			&device.DeviceType,
			&device.IsActive,
			&device.AppVersion,
			&device.OSVersion,
			&device.LastSeenAt,
			&device.CreatedAt,
			&device.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}
