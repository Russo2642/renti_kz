package utils

import (
	"github.com/russo2642/renti_kz/internal/domain"
)

const UserSelectFields = `
	u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
	u.role_id, u.is_active, u.password_hash, u.created_at, u.updated_at, r.name`

const UserSelectFieldsNoRole = `
	id, phone, first_name, last_name, email, city_id, iin, 
	role_id, is_active, password_hash, created_at, updated_at`

func ScanUser(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.User, error) {
	user := &domain.User{}

	err := scanner.Scan(
		&user.ID, &user.Phone, &user.FirstName, &user.LastName, &user.Email,
		&user.CityID, &user.IIN, &user.RoleID, &user.IsActive, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.Role,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func ScanUserNoRole(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.User, error) {
	user := &domain.User{}

	err := scanner.Scan(
		&user.ID, &user.Phone, &user.FirstName, &user.LastName, &user.Email,
		&user.CityID, &user.IIN, &user.RoleID, &user.IsActive, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func LoadUserCity(user *domain.User, locationRepo domain.LocationRepository) {
	if user.CityID > 0 {
		city, err := locationRepo.GetCityByID(user.CityID)
		if err == nil {
			user.City = city
		}
	}
}

func LoadUserRole(user *domain.User, roleRepo domain.RoleRepository) {
	if user.RoleID > 0 {
		role, err := roleRepo.GetByID(user.RoleID)
		if err == nil {
			user.Role = domain.UserRole(role.Name)
		}
	}
}

func ResolveUserRoleID(user *domain.User, roleRepo domain.RoleRepository) error {
	if user.RoleID == 0 && user.Role != "" {
		role, err := roleRepo.GetByName(string(user.Role))
		if err != nil {
			return HandleSQLError(err, "role", "get by name")
		}
		user.RoleID = role.ID
	}
	return nil
}

func BuildUserQuery(whereClause string) string {
	query := `
		SELECT ` + UserSelectFields + `
		FROM users u
		JOIN user_roles r ON u.role_id = r.id`
	
	if whereClause != "" {
		query += `
		WHERE ` + whereClause
	}
	
	return query
}

const PropertyOwnerSelectFields = `
	id, user_id, created_at, updated_at`

func ScanPropertyOwner(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.PropertyOwner, error) {
	owner := &domain.PropertyOwner{}

	err := scanner.Scan(
		&owner.ID, &owner.UserID, &owner.CreatedAt, &owner.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return owner, nil
}

const RenterSelectFields = `
	id, user_id, created_at, updated_at`

func ScanRenter(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Renter, error) {
	renter := &domain.Renter{}

	err := scanner.Scan(
		&renter.ID, &renter.UserID, &renter.CreatedAt, &renter.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return renter, nil
}
