package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type PropertyOwnerRepository struct {
	db *sql.DB
}

func NewPropertyOwnerRepository(db *sql.DB) *PropertyOwnerRepository {
	return &PropertyOwnerRepository{
		db: db,
	}
}

func (r *PropertyOwnerRepository) Create(owner *domain.PropertyOwner) error {
	now := time.Now()
	owner.CreatedAt = now
	owner.UpdatedAt = now

	query := `
		INSERT INTO property_owners (user_id, created_at, updated_at) 
		VALUES ($1, $2, $3) 
		RETURNING id`

	err := r.db.QueryRow(
		query,
		owner.UserID, owner.CreatedAt, owner.UpdatedAt,
	).Scan(&owner.ID)

	if err != nil {
		return utils.HandleSQLError(err, "property owner", "create")
	}

	return nil
}

func (r *PropertyOwnerRepository) GetByID(id int) (*domain.PropertyOwner, error) {
	query := `SELECT ` + utils.PropertyOwnerSelectFields + ` FROM property_owners WHERE id = $1`

	owner, err := utils.ScanPropertyOwner(r.db.QueryRow(query, id))
	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "property owner", "get", id)
	}

	return owner, nil
}

func (r *PropertyOwnerRepository) GetByUserID(userID int) (*domain.PropertyOwner, error) {
	query := `SELECT ` + utils.PropertyOwnerSelectFields + ` FROM property_owners WHERE user_id = $1`

	owner, err := utils.ScanPropertyOwner(r.db.QueryRow(query, userID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "property owner by user", "get")
	}

	return owner, nil
}

func (r *PropertyOwnerRepository) GetByUserIDWithUser(userID int) (*domain.PropertyOwner, error) {
	owner := &domain.PropertyOwner{}
	owner.User = &domain.User{}

	query := `
		SELECT 
			po.id, po.user_id, po.created_at, po.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
			u.role_id, u.password_hash, u.created_at, u.updated_at, ur.name
		FROM property_owners po
		JOIN users u ON po.user_id = u.id
		JOIN user_roles ur ON u.role_id = ur.id
		WHERE po.user_id = $1`

	err := r.db.QueryRow(query, userID).Scan(
		&owner.ID, &owner.UserID, &owner.CreatedAt, &owner.UpdatedAt,
		&owner.User.ID, &owner.User.Phone, &owner.User.FirstName, &owner.User.LastName, &owner.User.Email,
		&owner.User.CityID, &owner.User.IIN, &owner.User.RoleID, &owner.User.PasswordHash,
		&owner.User.CreatedAt, &owner.User.UpdatedAt, &owner.User.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "property owner with user", "get")
	}

	return owner, nil
}

func (r *PropertyOwnerRepository) GetByIDWithUser(id int) (*domain.PropertyOwner, error) {
	owner := &domain.PropertyOwner{}
	owner.User = &domain.User{}

	query := `
		SELECT 
			po.id, po.user_id, po.created_at, po.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
			u.role_id, u.password_hash, u.created_at, u.updated_at, ur.name
		FROM property_owners po
		JOIN users u ON po.user_id = u.id
		JOIN user_roles ur ON u.role_id = ur.id
		WHERE po.id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&owner.ID, &owner.UserID, &owner.CreatedAt, &owner.UpdatedAt,
		&owner.User.ID, &owner.User.Phone, &owner.User.FirstName, &owner.User.LastName, &owner.User.Email,
		&owner.User.CityID, &owner.User.IIN, &owner.User.RoleID, &owner.User.PasswordHash,
		&owner.User.CreatedAt, &owner.User.UpdatedAt, &owner.User.Role,
	)

	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "property owner with user", "get", id)
	}

	return owner, nil
}

func (r *PropertyOwnerRepository) Update(owner *domain.PropertyOwner) error {
	owner.UpdatedAt = time.Now()

	query := `
		UPDATE property_owners
		SET 
			updated_at = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(
		query,
		owner.ID, owner.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update property owner: %w", err)
	}

	return nil
}

func (r *PropertyOwnerRepository) Delete(id int) error {
	query := `DELETE FROM property_owners WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete property owner: %w", err)
	}

	return nil
}
