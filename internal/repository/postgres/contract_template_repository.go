package postgres

import (
	"database/sql"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type contractTemplateRepository struct {
	db *sql.DB
}

func NewContractTemplateRepository(db *sql.DB) domain.ContractTemplateRepository {
	return &contractTemplateRepository{db: db}
}

func (r *contractTemplateRepository) GetByTypeAndVersion(contractType domain.ContractType, version int) (*domain.ContractTemplate, error) {
	query := `
		SELECT id, type, version, template_content, is_active, created_at
		FROM contract_templates
		WHERE type = $1 AND version = $2`

	template := &domain.ContractTemplate{}
	err := r.db.QueryRow(query, contractType, version).Scan(
		&template.ID,
		&template.Type,
		&template.Version,
		&template.TemplateContent,
		&template.IsActive,
		&template.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template for type %s version %d not found", contractType, version)
		}
		return nil, utils.HandleSQLError(err, "contract template", "get")
	}

	return template, nil
}

func (r *contractTemplateRepository) GetActiveByType(contractType domain.ContractType) (*domain.ContractTemplate, error) {
	query := `
		SELECT id, type, version, template_content, is_active, created_at
		FROM contract_templates
		WHERE type = $1 AND is_active = true
		ORDER BY version DESC
		LIMIT 1`

	template := &domain.ContractTemplate{}
	err := r.db.QueryRow(query, contractType).Scan(
		&template.ID,
		&template.Type,
		&template.Version,
		&template.TemplateContent,
		&template.IsActive,
		&template.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active template found for type %s", contractType)
		}
		return nil, utils.HandleSQLError(err, "active contract template", "get")
	}

	return template, nil
}

func (r *contractTemplateRepository) GetLatestVersion(contractType domain.ContractType) (int, error) {
	query := `
		SELECT COALESCE(MAX(version), 0)
		FROM contract_templates
		WHERE type = $1`

	var version int
	err := r.db.QueryRow(query, contractType).Scan(&version)
	if err != nil {
		return 0, utils.HandleSQLError(err, "latest template version", "get")
	}

	return version, nil
}

func (r *contractTemplateRepository) Create(template *domain.ContractTemplate) error {
	query := `
		INSERT INTO contract_templates (
			type, version, template_content, is_active
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id, created_at`

	err := r.db.QueryRow(
		query,
		template.Type,
		template.Version,
		template.TemplateContent,
		template.IsActive,
	).Scan(&template.ID, &template.CreatedAt)

	if err != nil {
		return utils.HandleSQLError(err, "contract template", "create")
	}

	return nil
}

func (r *contractTemplateRepository) Update(template *domain.ContractTemplate) error {
	query := `
		UPDATE contract_templates SET 
			template_content = $2,
			is_active = $3
		WHERE id = $1`

	result, err := r.db.Exec(
		query,
		template.ID,
		template.TemplateContent,
		template.IsActive,
	)

	if err != nil {
		return utils.HandleSQLError(err, "contract template", "update")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleSQLError(err, "contract template update", "check rows affected")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contract template with id %d not found", template.ID)
	}

	return nil
}

func (r *contractTemplateRepository) SetActiveVersion(contractType domain.ContractType, version int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return utils.HandleSQLError(err, "contract template", "begin transaction")
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE contract_templates SET is_active = false WHERE type = $1`,
		contractType,
	)
	if err != nil {
		return utils.HandleSQLError(err, "contract template", "deactivate all")
	}

	result, err := tx.Exec(
		`UPDATE contract_templates SET is_active = true WHERE type = $1 AND version = $2`,
		contractType, version,
	)
	if err != nil {
		return utils.HandleSQLError(err, "contract template", "activate version")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleSQLError(err, "contract template activation", "check rows affected")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template for type %s version %d not found", contractType, version)
	}

	if err = tx.Commit(); err != nil {
		return utils.HandleSQLError(err, "contract template", "commit transaction")
	}

	return nil
}
