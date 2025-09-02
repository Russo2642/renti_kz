package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type contractRepository struct {
	db *sql.DB
}

func NewContractRepository(db *sql.DB) domain.ContractRepository {
	return &contractRepository{db: db}
}

func (r *contractRepository) Create(contract *domain.Contract) error {
	query := `
		INSERT INTO contracts (
			type, apartment_id, booking_id, template_version, 
			data_snapshot, status, is_active, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id, created_at, updated_at`

	var dataSnapshot []byte
	if contract.DataSnapshot != nil {
		dataSnapshot = contract.DataSnapshot
	}

	err := r.db.QueryRow(
		query,
		contract.Type,
		contract.ApartmentID,
		contract.BookingID,
		contract.TemplateVersion,
		dataSnapshot,
		contract.Status,
		contract.IsActive,
		contract.ExpiresAt,
	).Scan(&contract.ID, &contract.CreatedAt, &contract.UpdatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "contract", "create")
	}

	return nil
}

func (r *contractRepository) GetByID(id int) (*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		WHERE id = $1`

	contract := &domain.Contract{}
	var dataSnapshot []byte

	err := r.db.QueryRow(query, id).Scan(
		&contract.ID,
		&contract.Type,
		&contract.ApartmentID,
		&contract.BookingID,
		&contract.TemplateVersion,
		&dataSnapshot,
		&contract.Status,
		&contract.IsActive,
		&contract.ExpiresAt,
		&contract.CreatedAt,
		&contract.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contract with id %d not found", id)
		}
		return nil, utils.HandleSQLErrorWithID(err, "contract", "get", id)
	}

	if len(dataSnapshot) > 0 {
		contract.DataSnapshot = dataSnapshot
	}

	return contract, nil
}

func (r *contractRepository) Update(contract *domain.Contract) error {
	query := `
		UPDATE contracts SET 
			type = $2, 
			apartment_id = $3,
			booking_id = $4,
			template_version = $5, 
			data_snapshot = $6,
			status = $7,
			is_active = $8,
			expires_at = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	var dataSnapshot []byte
	if contract.DataSnapshot != nil {
		dataSnapshot = contract.DataSnapshot
	}

	result, err := r.db.Exec(
		query,
		contract.ID,
		contract.Type,
		contract.ApartmentID,
		contract.BookingID,
		contract.TemplateVersion,
		dataSnapshot,
		contract.Status,
		contract.IsActive,
		contract.ExpiresAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "contract", "update")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleSQLError(err, "contract update", "check rows affected")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contract with id %d not found", contract.ID)
	}

	return nil
}

func (r *contractRepository) Delete(id int) error {
	query := `DELETE FROM contracts WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLError(err, "contract", "delete")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleSQLError(err, "contract delete", "check rows affected")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contract with id %d not found", id)
	}

	return nil
}

func (r *contractRepository) GetByBookingID(bookingID int) (*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		WHERE booking_id = $1`

	contract := &domain.Contract{}
	var dataSnapshot []byte

	err := r.db.QueryRow(query, bookingID).Scan(
		&contract.ID,
		&contract.Type,
		&contract.ApartmentID,
		&contract.BookingID,
		&contract.TemplateVersion,
		&dataSnapshot,
		&contract.Status,
		&contract.IsActive,
		&contract.ExpiresAt,
		&contract.CreatedAt,
		&contract.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contract for booking %d not found", bookingID)
		}
		return nil, utils.HandleSQLError(err, "contract by booking", "get")
	}

	if len(dataSnapshot) > 0 {
		contract.DataSnapshot = dataSnapshot
	}

	return contract, nil
}

func (r *contractRepository) GetContractIDByBookingID(bookingID int) (*int, error) {
	query := `SELECT id FROM contracts WHERE booking_id = $1`

	var contractID int
	err := r.db.QueryRow(query, bookingID).Scan(&contractID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "contract id by booking", "get")
	}

	return &contractID, nil
}

func (r *contractRepository) GetByApartmentID(apartmentID int, contractType domain.ContractType) (*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		WHERE apartment_id = $1 AND type = $2 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1`

	contract := &domain.Contract{}
	var dataSnapshot []byte

	err := r.db.QueryRow(query, apartmentID, contractType).Scan(
		&contract.ID,
		&contract.Type,
		&contract.ApartmentID,
		&contract.BookingID,
		&contract.TemplateVersion,
		&dataSnapshot,
		&contract.Status,
		&contract.IsActive,
		&contract.ExpiresAt,
		&contract.CreatedAt,
		&contract.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contract for apartment %d with type %s not found", apartmentID, contractType)
		}
		return nil, utils.HandleSQLError(err, "contract by apartment", "get")
	}

	if len(dataSnapshot) > 0 {
		contract.DataSnapshot = dataSnapshot
	}

	return contract, nil
}

func (r *contractRepository) GetByApartmentIDAndType(apartmentID int, contractType domain.ContractType) ([]*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		WHERE apartment_id = $1 AND type = $2
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, apartmentID, contractType)
	if err != nil {
		return nil, utils.HandleSQLError(err, "contracts by apartment and type", "get")
	}
	defer utils.CloseRows(rows)

	return r.scanContracts(rows)
}

func (r *contractRepository) GetActiveByApartmentID(apartmentID int) ([]*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		WHERE apartment_id = $1 AND is_active = true
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "active contracts by apartment", "get")
	}
	defer utils.CloseRows(rows)

	return r.scanContracts(rows)
}

func (r *contractRepository) GetByStatus(status domain.ContractStatus, limit, offset int) ([]*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, status, limit, offset)
	if err != nil {
		return nil, utils.HandleSQLError(err, "contracts by status", "get")
	}
	defer utils.CloseRows(rows)

	return r.scanContracts(rows)
}

func (r *contractRepository) GetByType(contractType domain.ContractType, limit, offset int) ([]*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		WHERE type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, contractType, limit, offset)
	if err != nil {
		return nil, utils.HandleSQLError(err, "contracts by type", "get")
	}
	defer utils.CloseRows(rows)

	return r.scanContracts(rows)
}

func (r *contractRepository) GetAll(limit, offset int) ([]*domain.Contract, error) {
	query := `
		SELECT id, type, apartment_id, booking_id, template_version, 
			   data_snapshot, status, is_active, expires_at, created_at, updated_at
		FROM contracts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, utils.HandleSQLError(err, "contracts", "get all")
	}
	defer utils.CloseRows(rows)

	return r.scanContracts(rows)
}

func (r *contractRepository) ExistsForBooking(bookingID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM contracts WHERE booking_id = $1)`

	var exists bool
	err := r.db.QueryRow(query, bookingID).Scan(&exists)
	if err != nil {
		return false, utils.HandleSQLError(err, "contract existence", "check")
	}

	return exists, nil
}

func (r *contractRepository) CountByType(contractType domain.ContractType) (int, error) {
	query := `SELECT COUNT(*) FROM contracts WHERE type = $1`

	var count int
	err := r.db.QueryRow(query, contractType).Scan(&count)
	if err != nil {
		return 0, utils.HandleSQLError(err, "contract count", "get")
	}

	return count, nil
}

func (r *contractRepository) scanContracts(rows *sql.Rows) ([]*domain.Contract, error) {
	var contracts []*domain.Contract

	for rows.Next() {
		contract := &domain.Contract{}
		var dataSnapshot []byte

		err := rows.Scan(
			&contract.ID,
			&contract.Type,
			&contract.ApartmentID,
			&contract.BookingID,
			&contract.TemplateVersion,
			&dataSnapshot,
			&contract.Status,
			&contract.IsActive,
			&contract.ExpiresAt,
			&contract.CreatedAt,
			&contract.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "contract", "scan")
		}

		if len(dataSnapshot) > 0 {
			contract.DataSnapshot = dataSnapshot
		}

		contracts = append(contracts, contract)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.HandleSQLError(err, "contracts", "iterate")
	}

	return contracts, nil
}

func (r *contractRepository) GetContractWithSnapshot(contractID int) (*domain.Contract, *domain.RentalContractSnapshot, error) {
	contract, err := r.GetByID(contractID)
	if err != nil {
		return nil, nil, err
	}

	if contract.DataSnapshot == nil {
		return contract, nil, nil
	}

	var snapshot domain.RentalContractSnapshot
	err = json.Unmarshal(contract.DataSnapshot, &snapshot)
	if err != nil {
		return contract, nil, fmt.Errorf("ошибка декодирования снапшота: %w", err)
	}

	return contract, &snapshot, nil
}
