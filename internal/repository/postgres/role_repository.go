package postgres

import (
	"database/sql"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

func (r *RoleRepository) GetAll() ([]*domain.Role, error) {
	query := `
		SELECT 
			id, name, description, created_at
		FROM user_roles
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over roles: %w", err)
	}

	return roles, nil
}

func (r *RoleRepository) GetByID(id int) (*domain.Role, error) {
	role := &domain.Role{}

	query := `
		SELECT 
			id, name, description, created_at
		FROM user_roles
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	return role, nil
}

func (r *RoleRepository) GetByName(name string) (*domain.Role, error) {
	role := &domain.Role{}

	query := `
		SELECT 
			id, name, description, created_at
		FROM user_roles
		WHERE name = $1
	`

	err := r.db.QueryRow(query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role with name %s not found", name)
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	return role, nil
}
