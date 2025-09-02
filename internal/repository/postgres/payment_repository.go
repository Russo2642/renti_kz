package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/russo2642/renti_kz/internal/domain"
)

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) domain.PaymentRepository {
	return &paymentRepository{
		db: db,
	}
}

func (r *paymentRepository) Create(payment *domain.Payment) error {
	var providerResponseJSON interface{}
	var err error

	if payment.ProviderResponse != nil {
		jsonBytes, err := json.Marshal(payment.ProviderResponse)
		if err != nil {
			return fmt.Errorf("failed to marshal provider response: %w", err)
		}
		providerResponseJSON = string(jsonBytes)
	} else {
		providerResponseJSON = nil
	}

	query := `
		INSERT INTO payments (
			booking_id, payment_id, amount, currency, status,
			payment_method, provider_status, provider_response,
			final_booking_status, processed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err = r.db.QueryRow(
		query,
		payment.BookingID,
		payment.PaymentID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.ProviderStatus,
		providerResponseJSON,
		payment.FinalBookingStatus,
		payment.ProcessedAt,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		fmt.Printf("❌ Payment creation failed: %v\n", err)
		return fmt.Errorf("database error creating payment: %w", err)
	}

	fmt.Printf("✅ Payment created successfully with ID: %d\n", payment.ID)
	return nil
}

func (r *paymentRepository) GetByID(id int64) (*domain.Payment, error) {
	payment := &domain.Payment{}
	var providerResponseJSON []byte

	query := `
		SELECT id, booking_id, payment_id, amount, currency, status,
			   payment_method, provider_status, provider_response,
			   final_booking_status, processed_at, created_at, updated_at
		FROM payments
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&payment.ID,
		&payment.BookingID,
		&payment.PaymentID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.ProviderStatus,
		&providerResponseJSON,
		&payment.FinalBookingStatus,
		&payment.ProcessedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if len(providerResponseJSON) > 0 {
		var providerResponse domain.FreedomPayStatusResponse
		if err := json.Unmarshal(providerResponseJSON, &providerResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider response: %w", err)
		}
		payment.ProviderResponse = &providerResponse
	}

	return payment, nil
}

func (r *paymentRepository) GetByPaymentID(paymentID string) (*domain.Payment, error) {
	payment := &domain.Payment{}
	var providerResponseJSON []byte

	query := `
		SELECT id, booking_id, payment_id, amount, currency, status,
			   payment_method, provider_status, provider_response,
			   final_booking_status, processed_at, created_at, updated_at
		FROM payments
		WHERE payment_id = $1`

	err := r.db.QueryRow(query, paymentID).Scan(
		&payment.ID,
		&payment.BookingID,
		&payment.PaymentID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.ProviderStatus,
		&providerResponseJSON,
		&payment.FinalBookingStatus,
		&payment.ProcessedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if len(providerResponseJSON) > 0 {
		var providerResponse domain.FreedomPayStatusResponse
		if err := json.Unmarshal(providerResponseJSON, &providerResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider response: %w", err)
		}
		payment.ProviderResponse = &providerResponse
	}

	return payment, nil
}

func (r *paymentRepository) GetByBookingID(bookingID int64) ([]*domain.Payment, error) {
	query := `
		SELECT id, booking_id, payment_id, amount, currency, status,
			   payment_method, provider_status, provider_response,
			   final_booking_status, processed_at, created_at, updated_at
		FROM payments
		WHERE booking_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		var providerResponseJSON []byte

		err := rows.Scan(
			&payment.ID,
			&payment.BookingID,
			&payment.PaymentID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.PaymentMethod,
			&payment.ProviderStatus,
			&providerResponseJSON,
			&payment.FinalBookingStatus,
			&payment.ProcessedAt,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(providerResponseJSON) > 0 {
			var providerResponse domain.FreedomPayStatusResponse
			if err := json.Unmarshal(providerResponseJSON, &providerResponse); err != nil {
				return nil, fmt.Errorf("failed to unmarshal provider response: %w", err)
			}
			payment.ProviderResponse = &providerResponse
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *paymentRepository) Update(payment *domain.Payment) error {
	var providerResponseJSON interface{}
	var err error

	if payment.ProviderResponse != nil {
		jsonBytes, err := json.Marshal(payment.ProviderResponse)
		if err != nil {
			return fmt.Errorf("failed to marshal provider response: %w", err)
		}
		providerResponseJSON = string(jsonBytes)
	} else {
		providerResponseJSON = nil
	}

	query := `
		UPDATE payments SET
			status = $2,
			payment_method = $3,
			provider_status = $4,
			provider_response = $5,
			final_booking_status = $6,
			processed_at = $7
		WHERE id = $1
		RETURNING updated_at`

	err = r.db.QueryRow(
		query,
		payment.ID,
		payment.Status,
		payment.PaymentMethod,
		payment.ProviderStatus,
		providerResponseJSON,
		payment.FinalBookingStatus,
		payment.ProcessedAt,
	).Scan(&payment.UpdatedAt)

	return err
}

func (r *paymentRepository) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.Payment, int, error) {
	whereClause, args := r.buildWhereClause(filters)

	countQuery := "SELECT COUNT(*) FROM payments" + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, booking_id, payment_id, amount, currency, status,
			   payment_method, provider_status, provider_response,
			   final_booking_status, processed_at, created_at, updated_at
		FROM payments` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		var providerResponseJSON []byte

		err := rows.Scan(
			&payment.ID,
			&payment.BookingID,
			&payment.PaymentID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.PaymentMethod,
			&payment.ProviderStatus,
			&providerResponseJSON,
			&payment.FinalBookingStatus,
			&payment.ProcessedAt,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(providerResponseJSON) > 0 {
			var providerResponse domain.FreedomPayStatusResponse
			if err := json.Unmarshal(providerResponseJSON, &providerResponse); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal provider response: %w", err)
			}
			payment.ProviderResponse = &providerResponse
		}

		payments = append(payments, payment)
	}

	return payments, total, nil
}

func (r *paymentRepository) buildWhereClause(filters map[string]interface{}) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "status":
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
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
		}
	}

	if len(conditions) > 0 {
		return " WHERE " + strings.Join(conditions, " AND "), args
	}

	return "", nil
}
