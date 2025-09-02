package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type UserRepository struct {
	db           *sql.DB
	roleRepo     domain.RoleRepository
	locationRepo domain.LocationRepository
}

func NewUserRepository(db *sql.DB, roleRepo domain.RoleRepository, locationRepo domain.LocationRepository) *UserRepository {
	return &UserRepository{
		db:           db,
		roleRepo:     roleRepo,
		locationRepo: locationRepo,
	}
}

func (r *UserRepository) Create(user *domain.User) error {

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if !user.IsActive {
		user.IsActive = true
	}

	if err := utils.ResolveUserRoleID(user, r.roleRepo); err != nil {
		return err
	}

	query := `
		INSERT INTO users (
			phone, first_name, last_name, email, city_id, iin, role_id, 
			is_active, password_hash, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		RETURNING id`

	err := r.db.QueryRow(
		query,
		user.Phone, user.FirstName, user.LastName, user.Email,
		user.CityID, user.IIN, user.RoleID, user.IsActive, user.PasswordHash,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return utils.HandleSQLError(err, "user", "create")
	}

	utils.LoadUserRole(user, r.roleRepo)

	return nil
}

func (r *UserRepository) GetByID(id int) (*domain.User, error) {
	query := utils.BuildUserQuery("u.id = $1")

	user, err := utils.ScanUser(r.db.QueryRow(query, id))
	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "user", "get", id)
	}

	utils.LoadUserCity(user, r.locationRepo)

	return user, nil
}

func (r *UserRepository) GetByPhone(phone string) (*domain.User, error) {
	query := utils.BuildUserQuery("u.phone = $1")

	user, err := utils.ScanUser(r.db.QueryRow(query, phone))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "user by phone", "get")
	}

	utils.LoadUserCity(user, r.locationRepo)

	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	query := utils.BuildUserQuery("u.email = $1")

	user, err := utils.ScanUser(r.db.QueryRow(query, email))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "user by email", "get")
	}

	utils.LoadUserCity(user, r.locationRepo)

	return user, nil
}

func (r *UserRepository) Update(user *domain.User) error {

	user.UpdatedAt = time.Now()

	if err := utils.ResolveUserRoleID(user, r.roleRepo); err != nil {
		return err
	}

	query := `
		UPDATE users
		SET 
			phone = $2, first_name = $3, last_name = $4, email = $5,
			city_id = $6, iin = $7, role_id = $8, is_active = $9, 
			password_hash = $10, updated_at = $11
		WHERE id = $1`

	_, err := r.db.Exec(
		query,
		user.ID, user.Phone, user.FirstName, user.LastName, user.Email,
		user.CityID, user.IIN, user.RoleID, user.IsActive, user.PasswordHash,
		user.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLErrorWithID(err, "user", "update", user.ID)
	}

	utils.LoadUserRole(user, r.roleRepo)

	utils.LoadUserCity(user, r.locationRepo)

	return nil
}

func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "user", "delete", id)
	}

	return nil
}

func (r *UserRepository) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.User, int, error) {
	var conditions []string
	args := []interface{}{}
	argCount := 0

	if role, ok := filters["role"].(domain.UserRole); ok && role != "" {
		conditions = append(conditions, "r.name = $"+strconv.Itoa(argCount+1))
		args = append(args, string(role))
		argCount++
	}

	if cityID, ok := filters["city_id"].(int); ok && cityID > 0 {
		conditions = append(conditions, "u.city_id = $"+strconv.Itoa(argCount+1))
		args = append(args, cityID)
		argCount++
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		searchParam := "$" + strconv.Itoa(argCount+1)
		condition := "(u.first_name ILIKE " + searchParam +
			" OR u.last_name ILIKE " + searchParam +
			" OR u.phone ILIKE " + searchParam +
			" OR u.email ILIKE " + searchParam + ")"
		conditions = append(conditions, condition)
		args = append(args, "%"+search+"%")
		argCount++
	}

	if verificationStatus, ok := filters["verification_status"].(string); ok && verificationStatus != "" {
		conditions = append(conditions, "rnt.verification_status = $"+strconv.Itoa(argCount+1))
		args = append(args, verificationStatus)
		argCount++
	}

	if isActive, ok := filters["is_active"].(bool); ok {
		conditions = append(conditions, "u.is_active = $"+strconv.Itoa(argCount+1))
		args = append(args, isActive)
		argCount++
	}

	var whereClause string
	if len(conditions) > 0 {
		whereClause = conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	whereSQL := ""
	if whereClause != "" {
		whereSQL = " WHERE " + whereClause
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM users u
		JOIN user_roles r ON u.role_id = r.id
		LEFT JOIN renters rnt ON u.id = rnt.user_id%s`, whereSQL)

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "users count", "get")
	}

	offset := (page - 1) * pageSize
	limitClause := " LIMIT $" + strconv.Itoa(argCount+1) + " OFFSET $" + strconv.Itoa(argCount+2)
	args = append(args, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
			u.role_id, u.is_active, u.password_hash, u.created_at, u.updated_at, r.name
		FROM users u
		JOIN user_roles r ON u.role_id = r.id
		LEFT JOIN renters rnt ON u.id = rnt.user_id%s
		ORDER BY u.created_at DESC%s`, whereSQL, limitClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, utils.HandleSQLError(err, "users", "get")
	}
	defer utils.CloseRows(rows)

	var users []*domain.User
	for rows.Next() {
		user, err := utils.ScanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	for _, user := range users {
		utils.LoadUserCity(user, r.locationRepo)
	}

	return users, total, nil
}

func (r *UserRepository) UpdateRole(userID int, role domain.UserRole) error {
	roleEntity, err := r.roleRepo.GetByName(string(role))
	if err != nil {
		return err
	}

	query := `UPDATE users SET role_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = r.db.Exec(query, roleEntity.ID, userID)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "user role", "update", userID)
	}

	return nil
}

func (r *UserRepository) UpdateStatus(userID int, isActive bool) error {
	query := `UPDATE users SET is_active = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, isActive, userID)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "user status", "update", userID)
	}

	return nil
}

func (r *UserRepository) GetRoleStatistics() (map[string]int, error) {
	query := `
		SELECT 
			CASE 
				WHEN ur.name IN ('user', 'renter') THEN 'user'
				WHEN ur.name IN ('owner', 'property_owner') THEN 'owner'
				WHEN ur.name = 'moderator' THEN 'moderator'
				WHEN ur.name = 'admin' THEN 'admin'
				ELSE ur.name
			END as normalized_role,
			COUNT(*) as count
		FROM users u
		JOIN user_roles ur ON u.role_id = ur.id
		GROUP BY normalized_role`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get role statistics: %w", err)
	}
	defer rows.Close()

	statistics := make(map[string]int)
	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, fmt.Errorf("failed to scan role statistics: %w", err)
		}
		statistics[role] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return statistics, nil
}

func (r *UserRepository) GetStatusStatistics() (map[string]int, error) {
	query := `
		SELECT 
			CASE 
				WHEN is_active = true THEN 'active'
				ELSE 'inactive'
			END as status,
			COUNT(*) as count
		FROM users
		GROUP BY is_active`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}
	defer rows.Close()

	statistics := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status statistics: %w", err)
		}
		statistics[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return statistics, nil
}
