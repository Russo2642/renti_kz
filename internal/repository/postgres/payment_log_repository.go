package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/russo2642/renti_kz/internal/domain"
)

type paymentLogRepository struct {
	db *sql.DB
}

func NewPaymentLogRepository(db *sql.DB) domain.PaymentLogRepository {
	return &paymentLogRepository{
		db: db,
	}
}

func (r *paymentLogRepository) Create(log *domain.PaymentLog) error {
	var fpResponseJSON interface{}
	var err error

	if log.FPResponse != nil {
		jsonBytes, err := json.Marshal(log.FPResponse)
		if err != nil {
			return fmt.Errorf("failed to marshal fp response: %w", err)
		}
		fpResponseJSON = string(jsonBytes)
	} else {
		fpResponseJSON = nil
	}

	query := `
		INSERT INTO payment_logs (
			payment_id, booking_id, fp_payment_id, action, old_status, new_status,
			fp_response, processing_duration, user_id, source, success,
			error_message, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at`

	err = r.db.QueryRow(
		query,
		log.PaymentID,
		log.BookingID,
		log.FPPaymentID,
		log.Action,
		log.OldStatus,
		log.NewStatus,
		fpResponseJSON,
		log.ProcessingDuration,
		log.UserID,
		log.Source,
		log.Success,
		log.ErrorMessage,
		log.IPAddress,
		log.UserAgent,
	).Scan(&log.ID, &log.CreatedAt)

	return err
}

func (r *paymentLogRepository) GetByPaymentID(paymentID int64) ([]*domain.PaymentLog, error) {
	query := `
		SELECT id, payment_id, booking_id, fp_payment_id, action, old_status, new_status,
			   fp_response, processing_duration, user_id, source, success,
			   error_message, ip_address, user_agent, created_at
		FROM payment_logs
		WHERE payment_id = $1
		ORDER BY created_at DESC`

	return r.scanPaymentLogs(query, paymentID)
}

func (r *paymentLogRepository) GetByBookingID(bookingID int64) ([]*domain.PaymentLog, error) {
	query := `
		SELECT id, payment_id, booking_id, fp_payment_id, action, old_status, new_status,
			   fp_response, processing_duration, user_id, source, success,
			   error_message, ip_address, user_agent, created_at
		FROM payment_logs
		WHERE booking_id = $1
		ORDER BY created_at DESC`

	return r.scanPaymentLogs(query, bookingID)
}

func (r *paymentLogRepository) GetByFPPaymentID(fpPaymentID string) ([]*domain.PaymentLog, error) {
	query := `
		SELECT id, payment_id, booking_id, fp_payment_id, action, old_status, new_status,
			   fp_response, processing_duration, user_id, source, success,
			   error_message, ip_address, user_agent, created_at
		FROM payment_logs
		WHERE fp_payment_id = $1
		ORDER BY created_at DESC`

	return r.scanPaymentLogs(query, fpPaymentID)
}

func (r *paymentLogRepository) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.PaymentLog, int, error) {
	whereClause, args := r.buildWhereClause(filters)

	countQuery := "SELECT COUNT(*) FROM payment_logs" + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, payment_id, booking_id, fp_payment_id, action, old_status, new_status,
			   fp_response, processing_duration, user_id, source, success,
			   error_message, ip_address, user_agent, created_at
		FROM payment_logs` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, offset)

	logs, err := r.scanPaymentLogs(query, args...)
	return logs, total, err
}

func (r *paymentLogRepository) scanPaymentLogs(query string, args ...interface{}) ([]*domain.PaymentLog, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.PaymentLog
	for rows.Next() {
		log := &domain.PaymentLog{}
		var fpResponseJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.PaymentID,
			&log.BookingID,
			&log.FPPaymentID,
			&log.Action,
			&log.OldStatus,
			&log.NewStatus,
			&fpResponseJSON,
			&log.ProcessingDuration,
			&log.UserID,
			&log.Source,
			&log.Success,
			&log.ErrorMessage,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(fpResponseJSON) > 0 {
			var fpResponse domain.FreedomPayStatusResponse
			if err := json.Unmarshal(fpResponseJSON, &fpResponse); err != nil {
				return nil, fmt.Errorf("failed to unmarshal fp response: %w", err)
			}
			log.FPResponse = &fpResponse
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (r *paymentLogRepository) buildWhereClause(filters map[string]interface{}) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "action":
			conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "booking_id":
			conditions = append(conditions, fmt.Sprintf("booking_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "payment_id":
			conditions = append(conditions, fmt.Sprintf("payment_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "fp_payment_id":
			conditions = append(conditions, fmt.Sprintf("fp_payment_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "success":
			conditions = append(conditions, fmt.Sprintf("success = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "source":
			conditions = append(conditions, fmt.Sprintf("source = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "user_id":
			conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(conditions) > 0 {
		return " WHERE " + strings.Join(conditions, " AND "), args
	}

	return "", nil
}
