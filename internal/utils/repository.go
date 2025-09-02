package utils

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

func HandleSQLNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func StringToSQLNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func HandleSQLNullInt64(ni sql.NullInt64) *int {
	if ni.Valid {
		result := int(ni.Int64)
		return &result
	}
	return nil
}

func IntToSQLNullInt64(i *int) sql.NullInt64 {
	if i != nil {
		return sql.NullInt64{Int64: int64(*i), Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

func Int64ToSQLNullInt64(i *int64) sql.NullInt64 {
	if i != nil {
		return sql.NullInt64{Int64: *i, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

func IntToSQLNullInt32(i *int) sql.NullInt32 {
	if i != nil {
		return sql.NullInt32{Int32: int32(*i), Valid: true}
	}
	return sql.NullInt32{Valid: false}
}

func HandleSQLNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func TimeToSQLNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

func SetTimestamps(createdAt, updatedAt *time.Time) {
	now := time.Now()
	if createdAt != nil && createdAt.IsZero() {
		*createdAt = now
	}
	if updatedAt != nil {
		*updatedAt = now
	}
}

func CheckResourceExists(resource interface{}, resourceName string, id int) error {
	if resource == nil {
		return fmt.Errorf("%s with id %d not found", resourceName, id)
	}
	return nil
}

func HandleSQLError(err error, resourceName string, operation string) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return fmt.Errorf("%s not found", resourceName)
	}

	return fmt.Errorf("failed to %s %s: %w", operation, resourceName, err)
}

func HandleSQLErrorWithID(err error, resourceName string, operation string, id int) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return fmt.Errorf("%s with id %d not found", resourceName, id)
	}

	return fmt.Errorf("failed to %s %s with id %d: %w", operation, resourceName, id, err)
}

func BuildInClause(prefix string, items interface{}, startIndex int) (string, []interface{}) {
	switch v := items.(type) {
	case []string:
		if len(v) == 0 {
			return "", nil
		}
		placeholders := make([]string, len(v))
		args := make([]interface{}, len(v))
		for i, item := range v {
			placeholders[i] = fmt.Sprintf("$%d", startIndex+i)
			args[i] = item
		}
		return fmt.Sprintf("%s IN (%s)", prefix, strings.Join(placeholders, ",")), args

	case []int:
		if len(v) == 0 {
			return "", nil
		}
		placeholders := make([]string, len(v))
		args := make([]interface{}, len(v))
		for i, item := range v {
			placeholders[i] = fmt.Sprintf("$%d", startIndex+i)
			args[i] = item
		}
		return fmt.Sprintf("%s IN (%s)", prefix, strings.Join(placeholders, ",")), args

	default:
		return "", nil
	}
}

func CloseRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		log.Printf("Error closing rows: %v", err)
	}
}

func CheckRowsError(rows *sql.Rows, operation string) error {
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error during %s: %w", operation, err)
	}
	return nil
}

func BuildOrderByClause(orderBy, direction string, defaultOrder string) string {
	if orderBy == "" {
		return defaultOrder
	}

	if direction != "ASC" && direction != "DESC" {
		direction = "ASC"
	}

	return fmt.Sprintf("ORDER BY %s %s", orderBy, direction)
}

func BuildLimitOffsetClause(page, pageSize int) (string, []interface{}) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return "LIMIT $1 OFFSET $2", []interface{}{pageSize, offset}
}

func ScanNullableFields(dest ...*sql.NullString) []interface{} {
	result := make([]interface{}, len(dest))
	for i, d := range dest {
		result[i] = d
	}
	return result
}

func ExecuteInTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
