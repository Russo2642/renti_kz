package postgres

import (
	"database/sql"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
)

type cancellationRuleRepository struct {
	db *sql.DB
}

func NewCancellationRuleRepository(db *sql.DB) domain.CancellationRuleRepository {
	return &cancellationRuleRepository{db: db}
}

func (r *cancellationRuleRepository) GetAll() ([]*domain.CancellationRule, error) {
	query := `
		SELECT id, rule_type, title, content, is_active, display_order, created_at, updated_at
		FROM cancellation_rules
		WHERE is_active = true
		ORDER BY display_order ASC, created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var rules []*domain.CancellationRule
	for rows.Next() {
		rule := &domain.CancellationRule{}
		err := rows.Scan(
			&rule.ID,
			&rule.RuleType,
			&rule.Title,
			&rule.Content,
			&rule.IsActive,
			&rule.DisplayOrder,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки строк: %w", err)
	}

	return rules, nil
}

func (r *cancellationRuleRepository) GetByType(ruleType domain.CancellationRuleType) ([]*domain.CancellationRule, error) {
	query := `
		SELECT id, rule_type, title, content, is_active, display_order, created_at, updated_at
		FROM cancellation_rules
		WHERE rule_type = $1 AND is_active = true
		ORDER BY display_order ASC, created_at ASC
	`

	rows, err := r.db.Query(query, ruleType)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var rules []*domain.CancellationRule
	for rows.Next() {
		rule := &domain.CancellationRule{}
		err := rows.Scan(
			&rule.ID,
			&rule.RuleType,
			&rule.Title,
			&rule.Content,
			&rule.IsActive,
			&rule.DisplayOrder,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки строк: %w", err)
	}

	return rules, nil
}

func (r *cancellationRuleRepository) GetByID(id int) (*domain.CancellationRule, error) {
	query := `
		SELECT id, rule_type, title, content, is_active, display_order, created_at, updated_at
		FROM cancellation_rules
		WHERE id = $1
	`

	rule := &domain.CancellationRule{}
	err := r.db.QueryRow(query, id).Scan(
		&rule.ID,
		&rule.RuleType,
		&rule.Title,
		&rule.Content,
		&rule.IsActive,
		&rule.DisplayOrder,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения правила отмены: %w", err)
	}

	return rule, nil
}

func (r *cancellationRuleRepository) Create(rule *domain.CancellationRule) error {
	query := `
		INSERT INTO cancellation_rules (rule_type, title, content, is_active, display_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		rule.RuleType,
		rule.Title,
		rule.Content,
		rule.IsActive,
		rule.DisplayOrder,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("ошибка создания правила отмены: %w", err)
	}

	return nil
}

func (r *cancellationRuleRepository) Update(rule *domain.CancellationRule) error {
	query := `
		UPDATE cancellation_rules 
		SET rule_type = $2, title = $3, content = $4, is_active = $5, display_order = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		rule.ID,
		rule.RuleType,
		rule.Title,
		rule.Content,
		rule.IsActive,
		rule.DisplayOrder,
	).Scan(&rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("ошибка обновления правила отмены: %w", err)
	}

	return nil
}

func (r *cancellationRuleRepository) Delete(id int) error {
	query := `DELETE FROM cancellation_rules WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления правила отмены: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("правило отмены с ID %d не найдено", id)
	}

	return nil
}
