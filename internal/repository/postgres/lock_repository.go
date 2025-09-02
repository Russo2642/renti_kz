package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type lockRepository struct {
	db *sql.DB
}

func NewLockRepository(db *sql.DB) domain.LockRepository {
	return &lockRepository{db: db}
}

func (r *lockRepository) Create(lock *domain.Lock) error {
	query := `
		INSERT INTO locks (unique_id, apartment_id, name, description, current_status, firmware_version, tuya_device_id, owner_password)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		lock.UniqueID,
		lock.ApartmentID,
		lock.Name,
		lock.Description,
		lock.CurrentStatus,
		lock.FirmwareVersion,
		lock.TuyaDeviceID,
		lock.OwnerPassword,
	).Scan(&lock.ID, &lock.CreatedAt, &lock.UpdatedAt)

	return err
}

func (r *lockRepository) GetByID(id int) (*domain.Lock, error) {
	query := `
		SELECT id, unique_id, apartment_id, name, description, current_status,
		       last_status_update, last_heartbeat, is_online, firmware_version,
		       battery_level, signal_strength, 
		       COALESCE(tuya_device_id, '') as tuya_device_id, 
		       COALESCE(owner_password, '') as owner_password,
		       COALESCE(battery_type, 'unknown') as battery_type,
		       charging_status, last_battery_check, last_tuya_sync,
		       COALESCE(auto_update_enabled, false) as auto_update_enabled,
		       COALESCE(webhook_configured, false) as webhook_configured,
		       created_at, updated_at
		FROM locks
		WHERE id = $1`

	lock := &domain.Lock{}
	err := r.db.QueryRow(query, id).Scan(
		&lock.ID,
		&lock.UniqueID,
		&lock.ApartmentID,
		&lock.Name,
		&lock.Description,
		&lock.CurrentStatus,
		&lock.LastStatusUpdate,
		&lock.LastHeartbeat,
		&lock.IsOnline,
		&lock.FirmwareVersion,
		&lock.BatteryLevel,
		&lock.SignalStrength,
		&lock.TuyaDeviceID,
		&lock.OwnerPassword,
		&lock.BatteryType,
		&lock.ChargingStatus,
		&lock.LastBatteryCheck,
		&lock.LastTuyaSync,
		&lock.AutoUpdateEnabled,
		&lock.WebhookConfigured,
		&lock.CreatedAt,
		&lock.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return lock, nil
}

func (r *lockRepository) GetByUniqueID(uniqueID string) (*domain.Lock, error) {
	query := `
		SELECT id, unique_id, apartment_id, name, description, current_status,
		       last_status_update, last_heartbeat, is_online, firmware_version,
		       battery_level, signal_strength, 
		       COALESCE(tuya_device_id, '') as tuya_device_id, 
		       COALESCE(owner_password, '') as owner_password,
		       COALESCE(battery_type, 'unknown') as battery_type,
		       charging_status, last_battery_check, last_tuya_sync,
		       COALESCE(auto_update_enabled, false) as auto_update_enabled,
		       COALESCE(webhook_configured, false) as webhook_configured,
		       created_at, updated_at
		FROM locks
		WHERE unique_id = $1`

	lock := &domain.Lock{}
	err := r.db.QueryRow(query, uniqueID).Scan(
		&lock.ID,
		&lock.UniqueID,
		&lock.ApartmentID,
		&lock.Name,
		&lock.Description,
		&lock.CurrentStatus,
		&lock.LastStatusUpdate,
		&lock.LastHeartbeat,
		&lock.IsOnline,
		&lock.FirmwareVersion,
		&lock.BatteryLevel,
		&lock.SignalStrength,
		&lock.TuyaDeviceID,
		&lock.OwnerPassword,
		&lock.BatteryType,
		&lock.ChargingStatus,
		&lock.LastBatteryCheck,
		&lock.LastTuyaSync,
		&lock.AutoUpdateEnabled,
		&lock.WebhookConfigured,
		&lock.CreatedAt,
		&lock.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return lock, nil
}

func (r *lockRepository) GetByApartmentID(apartmentID int) (*domain.Lock, error) {
	query := `
		SELECT id, unique_id, apartment_id, name, description, current_status,
		       last_status_update, last_heartbeat, is_online, firmware_version,
		       battery_level, signal_strength, 
		       COALESCE(tuya_device_id, '') as tuya_device_id, 
		       COALESCE(owner_password, '') as owner_password,
		       COALESCE(battery_type, 'unknown') as battery_type,
		       charging_status, last_battery_check, last_tuya_sync,
		       COALESCE(auto_update_enabled, false) as auto_update_enabled,
		       COALESCE(webhook_configured, false) as webhook_configured,
		       created_at, updated_at
		FROM locks
		WHERE apartment_id = $1
		LIMIT 1`

	lock := &domain.Lock{}
	err := r.db.QueryRow(query, apartmentID).Scan(
		&lock.ID,
		&lock.UniqueID,
		&lock.ApartmentID,
		&lock.Name,
		&lock.Description,
		&lock.CurrentStatus,
		&lock.LastStatusUpdate,
		&lock.LastHeartbeat,
		&lock.IsOnline,
		&lock.FirmwareVersion,
		&lock.BatteryLevel,
		&lock.SignalStrength,
		&lock.TuyaDeviceID,
		&lock.OwnerPassword,
		&lock.BatteryType,
		&lock.ChargingStatus,
		&lock.LastBatteryCheck,
		&lock.LastTuyaSync,
		&lock.AutoUpdateEnabled,
		&lock.WebhookConfigured,
		&lock.CreatedAt,
		&lock.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return lock, nil
}

func (r *lockRepository) GetLocksByApartmentID(apartmentID int) ([]*domain.Lock, error) {
	query := `
		SELECT id, unique_id, apartment_id, name, description, current_status,
		       last_status_update, last_heartbeat, is_online, firmware_version,
		       battery_level, signal_strength, 
		       COALESCE(tuya_device_id, '') as tuya_device_id, 
		       COALESCE(owner_password, '') as owner_password,
		       COALESCE(battery_type, 'unknown') as battery_type,
		       charging_status, last_battery_check, last_tuya_sync,
		       COALESCE(auto_update_enabled, false) as auto_update_enabled,
		       COALESCE(webhook_configured, false) as webhook_configured,
		       created_at, updated_at
		FROM locks
		WHERE apartment_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []*domain.Lock
	for rows.Next() {
		lock := &domain.Lock{}
		err := rows.Scan(
			&lock.ID,
			&lock.UniqueID,
			&lock.ApartmentID,
			&lock.Name,
			&lock.Description,
			&lock.CurrentStatus,
			&lock.LastStatusUpdate,
			&lock.LastHeartbeat,
			&lock.IsOnline,
			&lock.FirmwareVersion,
			&lock.BatteryLevel,
			&lock.SignalStrength,
			&lock.TuyaDeviceID,
			&lock.OwnerPassword,
			&lock.BatteryType,
			&lock.ChargingStatus,
			&lock.LastBatteryCheck,
			&lock.LastTuyaSync,
			&lock.AutoUpdateEnabled,
			&lock.WebhookConfigured,
			&lock.CreatedAt,
			&lock.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locks = append(locks, lock)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return locks, nil
}

func (r *lockRepository) GetAll() ([]*domain.Lock, error) {
	query := `
		SELECT id, unique_id, apartment_id, name, description, current_status,
		       last_status_update, last_heartbeat, is_online, firmware_version,
		       battery_level, signal_strength, 
		       COALESCE(tuya_device_id, '') as tuya_device_id, 
		       COALESCE(owner_password, '') as owner_password,
		       COALESCE(battery_type, 'unknown') as battery_type,
		       charging_status, last_battery_check, last_tuya_sync,
		       COALESCE(auto_update_enabled, false) as auto_update_enabled,
		       COALESCE(webhook_configured, false) as webhook_configured,
		       created_at, updated_at
		FROM locks
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []*domain.Lock
	for rows.Next() {
		lock := &domain.Lock{}
		err := rows.Scan(
			&lock.ID,
			&lock.UniqueID,
			&lock.ApartmentID,
			&lock.Name,
			&lock.Description,
			&lock.CurrentStatus,
			&lock.LastStatusUpdate,
			&lock.LastHeartbeat,
			&lock.IsOnline,
			&lock.FirmwareVersion,
			&lock.BatteryLevel,
			&lock.SignalStrength,
			&lock.TuyaDeviceID,
			&lock.OwnerPassword,
			&lock.BatteryType,
			&lock.ChargingStatus,
			&lock.LastBatteryCheck,
			&lock.LastTuyaSync,
			&lock.AutoUpdateEnabled,
			&lock.WebhookConfigured,
			&lock.CreatedAt,
			&lock.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locks = append(locks, lock)
	}

	return locks, nil
}

func (r *lockRepository) GetAllWithFilters(filters map[string]interface{}, page, pageSize int) ([]*domain.Lock, int, error) {
	var conditions []string
	var args []interface{}
	argCount := 0

	if status, ok := filters["status"].(string); ok && status != "" {
		conditions = append(conditions, "current_status = $"+fmt.Sprintf("%d", argCount+1))
		args = append(args, status)
		argCount++
	}

	if minBatteryLevel, ok := filters["min_battery_level"].(int); ok && minBatteryLevel > 0 {
		conditions = append(conditions, "battery_level >= $"+fmt.Sprintf("%d", argCount+1))
		args = append(args, minBatteryLevel)
		argCount++
	}

	if maxBatteryLevel, ok := filters["max_battery_level"].(int); ok && maxBatteryLevel > 0 {
		conditions = append(conditions, "battery_level <= $"+fmt.Sprintf("%d", argCount+1))
		args = append(args, maxBatteryLevel)
		argCount++
	}

	if isOnline, ok := filters["is_online"].(bool); ok {
		conditions = append(conditions, "is_online = $"+fmt.Sprintf("%d", argCount+1))
		args = append(args, isOnline)
		argCount++
	}

	if apartmentID, ok := filters["apartment_id"].(int); ok && apartmentID > 0 {
		conditions = append(conditions, "apartment_id = $"+fmt.Sprintf("%d", argCount+1))
		args = append(args, apartmentID)
		argCount++
	}

	if unbound, ok := filters["unbound"].(bool); ok && unbound {
		conditions = append(conditions, "apartment_id IS NULL")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM locks %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	limitClause := fmt.Sprintf("LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT id, unique_id, apartment_id, name, description, current_status,
		       last_status_update, last_heartbeat, is_online, firmware_version,
		       battery_level, signal_strength, 
		       COALESCE(tuya_device_id, '') as tuya_device_id, 
		       COALESCE(owner_password, '') as owner_password,
		       COALESCE(battery_type, 'unknown') as battery_type,
		       charging_status, last_battery_check, last_tuya_sync,
		       COALESCE(auto_update_enabled, false) as auto_update_enabled,
		       COALESCE(webhook_configured, false) as webhook_configured,
		       created_at, updated_at
		FROM locks
		%s
		ORDER BY created_at DESC
		%s`, whereClause, limitClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var locks []*domain.Lock
	for rows.Next() {
		lock := &domain.Lock{}
		err := rows.Scan(
			&lock.ID,
			&lock.UniqueID,
			&lock.ApartmentID,
			&lock.Name,
			&lock.Description,
			&lock.CurrentStatus,
			&lock.LastStatusUpdate,
			&lock.LastHeartbeat,
			&lock.IsOnline,
			&lock.FirmwareVersion,
			&lock.BatteryLevel,
			&lock.SignalStrength,
			&lock.TuyaDeviceID,
			&lock.OwnerPassword,
			&lock.BatteryType,
			&lock.ChargingStatus,
			&lock.LastBatteryCheck,
			&lock.LastTuyaSync,
			&lock.AutoUpdateEnabled,
			&lock.WebhookConfigured,
			&lock.CreatedAt,
			&lock.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		locks = append(locks, lock)
	}

	return locks, total, nil
}

func (r *lockRepository) Update(lock *domain.Lock) error {
	query := `
		UPDATE locks SET
			apartment_id = $2, name = $3, description = $4, current_status = $5,
			last_status_update = $6, last_heartbeat = $7, is_online = $8,
			firmware_version = $9, battery_level = $10, signal_strength = $11,
			updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(
		query,
		lock.ID,
		lock.ApartmentID,
		lock.Name,
		lock.Description,
		lock.CurrentStatus,
		lock.LastStatusUpdate,
		lock.LastHeartbeat,
		lock.IsOnline,
		lock.FirmwareVersion,
		lock.BatteryLevel,
		lock.SignalStrength,
	)

	return err
}

func (r *lockRepository) Delete(id int) error {
	query := `DELETE FROM locks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *lockRepository) UpdateStatus(uniqueID string, status domain.LockStatus, timestamp *time.Time) error {
	var query string
	var args []interface{}

	if timestamp != nil {
		query = `
			UPDATE locks SET
				current_status = $2, last_status_update = $3, updated_at = NOW()
			WHERE unique_id = $1`
		args = []interface{}{uniqueID, status, timestamp}
	} else {
		query = `
			UPDATE locks SET
				current_status = $2, last_status_update = NOW(), updated_at = NOW()
			WHERE unique_id = $1`
		args = []interface{}{uniqueID, status}
	}

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *lockRepository) UpdateHeartbeat(uniqueID string, timestamp time.Time, batteryLevel *int, signalStrength *int) error {
	query := `
		UPDATE locks SET
			last_heartbeat = $2, battery_level = $3, signal_strength = $4,
			is_online = true, updated_at = NOW()
		WHERE unique_id = $1`

	_, err := r.db.Exec(query, uniqueID, timestamp, batteryLevel, signalStrength)
	return err
}

func (r *lockRepository) UpdateOnlineStatus(uniqueID string, isOnline bool) error {
	query := `UPDATE locks SET is_online = $2, updated_at = NOW() WHERE unique_id = $1`
	_, err := r.db.Exec(query, uniqueID, isOnline)
	return err
}

func (r *lockRepository) CreateStatusLog(log *domain.LockStatusLog) error {
	query := `
		INSERT INTO lock_status_logs (lock_id, old_status, new_status, change_source, user_id, booking_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.db.QueryRow(
		query,
		log.LockID,
		log.OldStatus,
		log.NewStatus,
		log.ChangeSource,
		log.UserID,
		log.BookingID,
		log.Notes,
	).Scan(&log.ID, &log.CreatedAt)

	return err
}

func (r *lockRepository) GetStatusLogsByLockID(lockID int, limit int) ([]*domain.LockStatusLog, error) {
	query := `
		SELECT id, lock_id, old_status, new_status, change_source, user_id, booking_id, notes, created_at
		FROM lock_status_logs
		WHERE lock_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(query, lockID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.LockStatusLog
	for rows.Next() {
		log := &domain.LockStatusLog{}
		err := rows.Scan(
			&log.ID,
			&log.LockID,
			&log.OldStatus,
			&log.NewStatus,
			&log.ChangeSource,
			&log.UserID,
			&log.BookingID,
			&log.Notes,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (r *lockRepository) GetStatusLogsByUniqueID(uniqueID string, limit int) ([]*domain.LockStatusLog, error) {
	query := `
		SELECT lsl.id, lsl.lock_id, lsl.old_status, lsl.new_status, lsl.change_source, 
		       lsl.user_id, lsl.booking_id, lsl.notes, lsl.created_at
		FROM lock_status_logs lsl
		JOIN locks l ON l.id = lsl.lock_id
		WHERE l.unique_id = $1
		ORDER BY lsl.created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(query, uniqueID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.LockStatusLog
	for rows.Next() {
		log := &domain.LockStatusLog{}
		err := rows.Scan(
			&log.ID,
			&log.LockID,
			&log.OldStatus,
			&log.NewStatus,
			&log.ChangeSource,
			&log.UserID,
			&log.BookingID,
			&log.Notes,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (r *lockRepository) CreateTempPassword(tempPassword *domain.LockTempPassword) error {
	query := `
		INSERT INTO lock_temp_passwords (lock_id, booking_id, user_id, password, tuya_password_id, name, valid_from, valid_until, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		tempPassword.LockID,
		tempPassword.BookingID,
		tempPassword.UserID,
		tempPassword.Password,
		tempPassword.TuyaPasswordID,
		tempPassword.Name,
		tempPassword.ValidFrom,
		tempPassword.ValidUntil,
		tempPassword.IsActive,
	).Scan(&tempPassword.ID, &tempPassword.CreatedAt, &tempPassword.UpdatedAt)

	return err
}

func (r *lockRepository) GetTempPasswordsByLockID(lockID int) ([]*domain.LockTempPassword, error) {
	query := `
		SELECT id, lock_id, booking_id, user_id, password, tuya_password_id, name, 
		       valid_from, valid_until, is_active, created_at, updated_at
		FROM lock_temp_passwords
		WHERE lock_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, lockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passwords []*domain.LockTempPassword
	for rows.Next() {
		password := &domain.LockTempPassword{}
		err := rows.Scan(
			&password.ID,
			&password.LockID,
			&password.BookingID,
			&password.UserID,
			&password.Password,
			&password.TuyaPasswordID,
			&password.Name,
			&password.ValidFrom,
			&password.ValidUntil,
			&password.IsActive,
			&password.CreatedAt,
			&password.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		passwords = append(passwords, password)
	}

	return passwords, nil
}

func (r *lockRepository) GetTempPasswordsByBookingID(bookingID int) ([]*domain.LockTempPassword, error) {
	query := `
		SELECT id, lock_id, booking_id, user_id, password, tuya_password_id, name, 
		       valid_from, valid_until, is_active, created_at, updated_at
		FROM lock_temp_passwords
		WHERE booking_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passwords []*domain.LockTempPassword
	for rows.Next() {
		password := &domain.LockTempPassword{}
		err := rows.Scan(
			&password.ID,
			&password.LockID,
			&password.BookingID,
			&password.UserID,
			&password.Password,
			&password.TuyaPasswordID,
			&password.Name,
			&password.ValidFrom,
			&password.ValidUntil,
			&password.IsActive,
			&password.CreatedAt,
			&password.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		passwords = append(passwords, password)
	}

	return passwords, nil
}

func (r *lockRepository) GetTempPasswordByID(id int) (*domain.LockTempPassword, error) {
	query := `
		SELECT id, lock_id, booking_id, user_id, password, tuya_password_id, name, 
		       valid_from, valid_until, is_active, created_at, updated_at
		FROM lock_temp_passwords
		WHERE id = $1`

	password := &domain.LockTempPassword{}
	err := r.db.QueryRow(query, id).Scan(
		&password.ID,
		&password.LockID,
		&password.BookingID,
		&password.UserID,
		&password.Password,
		&password.TuyaPasswordID,
		&password.Name,
		&password.ValidFrom,
		&password.ValidUntil,
		&password.IsActive,
		&password.CreatedAt,
		&password.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return password, nil
}

func (r *lockRepository) UpdateTempPassword(tempPassword *domain.LockTempPassword) error {
	query := `
		UPDATE lock_temp_passwords 
		SET password = $2, name = $3, valid_from = $4, valid_until = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(
		query,
		tempPassword.ID,
		tempPassword.Password,
		tempPassword.Name,
		tempPassword.ValidFrom,
		tempPassword.ValidUntil,
		tempPassword.IsActive,
	)

	return err
}

func (r *lockRepository) DeleteTempPassword(id int) error {
	query := `DELETE FROM lock_temp_passwords WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *lockRepository) DeactivateTempPassword(id int) error {
	query := `UPDATE lock_temp_passwords SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *lockRepository) UpdateBatteryInfo(uniqueID string, batteryLevel *int, batteryType domain.BatteryType, chargingStatus *domain.ChargingStatus) error {
	query := `
		UPDATE locks SET
			battery_level = $2, battery_type = $3, charging_status = $4,
			last_battery_check = NOW(), updated_at = NOW()
		WHERE unique_id = $1`

	_, err := r.db.Exec(query, uniqueID, batteryLevel, batteryType, chargingStatus)
	return err
}

func (r *lockRepository) UpdateTuyaSync(uniqueID string, syncTime time.Time) error {
	query := `UPDATE locks SET last_tuya_sync = $2, updated_at = NOW() WHERE unique_id = $1`
	_, err := r.db.Exec(query, uniqueID, syncTime)
	return err
}

func (r *lockRepository) EnableAutoUpdate(uniqueID string, enabled bool) error {
	query := `UPDATE locks SET auto_update_enabled = $2, updated_at = NOW() WHERE unique_id = $1`
	_, err := r.db.Exec(query, uniqueID, enabled)
	return err
}

func (r *lockRepository) ConfigureWebhook(uniqueID string, configured bool) error {
	query := `UPDATE locks SET webhook_configured = $2, updated_at = NOW() WHERE unique_id = $1`
	_, err := r.db.Exec(query, uniqueID, configured)
	return err
}

func (r *lockRepository) GetLocksForAutoUpdate() ([]*domain.Lock, error) {
	query := `
		SELECT id, unique_id, apartment_id, name, description, current_status,
		       last_status_update, last_heartbeat, is_online, firmware_version,
		       battery_level, signal_strength, 
		       COALESCE(tuya_device_id, '') as tuya_device_id, 
		       COALESCE(owner_password, '') as owner_password,
		       COALESCE(battery_type, 'unknown') as battery_type,
		       charging_status, last_battery_check, last_tuya_sync,
		       COALESCE(auto_update_enabled, false) as auto_update_enabled,
		       COALESCE(webhook_configured, false) as webhook_configured,
		       created_at, updated_at
		FROM locks
		WHERE auto_update_enabled = true
		ORDER BY last_tuya_sync ASC NULLS FIRST`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []*domain.Lock
	for rows.Next() {
		lock := &domain.Lock{}
		err := rows.Scan(
			&lock.ID,
			&lock.UniqueID,
			&lock.ApartmentID,
			&lock.Name,
			&lock.Description,
			&lock.CurrentStatus,
			&lock.LastStatusUpdate,
			&lock.LastHeartbeat,
			&lock.IsOnline,
			&lock.FirmwareVersion,
			&lock.BatteryLevel,
			&lock.SignalStrength,
			&lock.TuyaDeviceID,
			&lock.OwnerPassword,
			&lock.BatteryType,
			&lock.ChargingStatus,
			&lock.LastBatteryCheck,
			&lock.LastTuyaSync,
			&lock.AutoUpdateEnabled,
			&lock.WebhookConfigured,
			&lock.CreatedAt,
			&lock.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locks = append(locks, lock)
	}

	return locks, nil
}
