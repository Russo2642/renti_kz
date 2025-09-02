package postgres

import (
	"database/sql"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type apartmentTypeRepository struct {
	db *sql.DB
}

func NewApartmentTypeRepository(db *sql.DB) domain.ApartmentTypeRepository {
	return &apartmentTypeRepository{
		db: db,
	}
}

func (r *apartmentTypeRepository) Create(apartmentType *domain.ApartmentType) error {
	query := `
		INSERT INTO apartment_types (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		apartmentType.Name,
		apartmentType.Description,
	).Scan(
		&apartmentType.ID,
		&apartmentType.CreatedAt,
		&apartmentType.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLError(err, "apartment_type", "create")
	}

	return nil
}

func (r *apartmentTypeRepository) GetByID(id int) (*domain.ApartmentType, error) {
	apartmentType := &domain.ApartmentType{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM apartment_types
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&apartmentType.ID,
		&apartmentType.Name,
		&apartmentType.Description,
		&apartmentType.CreatedAt,
		&apartmentType.UpdatedAt,
	)

	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment_type", "get")
	}

	return apartmentType, nil
}

func (r *apartmentTypeRepository) GetAll() ([]*domain.ApartmentType, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM apartment_types
		ORDER BY name ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.HandleSQLError(err, "apartment_types", "get")
	}
	defer utils.CloseRows(rows)

	var apartmentTypes []*domain.ApartmentType
	for rows.Next() {
		apartmentType := &domain.ApartmentType{}
		err := rows.Scan(
			&apartmentType.ID,
			&apartmentType.Name,
			&apartmentType.Description,
			&apartmentType.CreatedAt,
			&apartmentType.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "apartment_types", "scan")
		}
		apartmentTypes = append(apartmentTypes, apartmentType)
	}

	if err = rows.Err(); err != nil {
		return nil, utils.HandleSQLError(err, "apartment_types", "iterate")
	}

	return apartmentTypes, nil
}

func (r *apartmentTypeRepository) Update(apartmentType *domain.ApartmentType) error {
	query := `
		UPDATE apartment_types 
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3`

	result, err := r.db.Exec(
		query,
		apartmentType.Name,
		apartmentType.Description,
		apartmentType.ID,
	)

	if err != nil {
		return utils.HandleSQLError(err, "apartment_type", "update")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleSQLError(err, "apartment_type", "update")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("apartment type with id %d not found", apartmentType.ID)
	}

	return nil
}

func (r *apartmentTypeRepository) Delete(id int) error {
	query := `DELETE FROM apartment_types WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLError(err, "apartment_type", "delete")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleSQLError(err, "apartment_type", "delete")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("apartment type with id %d not found", id)
	}

	return nil
}
