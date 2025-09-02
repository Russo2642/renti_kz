package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type platformSettingsRepository struct {
	db *sql.DB
}

func NewPlatformSettingsRepository(db *sql.DB) domain.PlatformSettingsRepository {
	return &platformSettingsRepository{
		db: db,
	}
}

func (r *platformSettingsRepository) GetByKey(key string) (*domain.PlatformSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description, data_type, is_active, created_at, updated_at
		FROM platform_settings 
		WHERE setting_key = $1 AND is_active = true
	`

	setting := &domain.PlatformSetting{}
	err := r.db.QueryRow(query, key).Scan(
		&setting.ID,
		&setting.SettingKey,
		&setting.SettingValue,
		&setting.Description,
		&setting.DataType,
		&setting.IsActive,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("настройка с ключом '%s' не найдена", key)
		}
		return nil, fmt.Errorf("ошибка получения настройки: %w", err)
	}

	return setting, nil
}

func (r *platformSettingsRepository) GetAll() ([]*domain.PlatformSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description, data_type, is_active, created_at, updated_at
		FROM platform_settings 
		ORDER BY setting_key
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех настроек: %w", err)
	}
	defer rows.Close()

	var settings []*domain.PlatformSetting
	for rows.Next() {
		setting := &domain.PlatformSetting{}
		err := rows.Scan(
			&setting.ID,
			&setting.SettingKey,
			&setting.SettingValue,
			&setting.Description,
			&setting.DataType,
			&setting.IsActive,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования настройки: %w", err)
		}
		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки результатов: %w", err)
	}

	return settings, nil
}

func (r *platformSettingsRepository) GetAllActive() ([]*domain.PlatformSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description, data_type, is_active, created_at, updated_at
		FROM platform_settings 
		WHERE is_active = true
		ORDER BY setting_key
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения активных настроек: %w", err)
	}
	defer rows.Close()

	var settings []*domain.PlatformSetting
	for rows.Next() {
		setting := &domain.PlatformSetting{}
		err := rows.Scan(
			&setting.ID,
			&setting.SettingKey,
			&setting.SettingValue,
			&setting.Description,
			&setting.DataType,
			&setting.IsActive,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования настройки: %w", err)
		}
		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки результатов: %w", err)
	}

	return settings, nil
}

func (r *platformSettingsRepository) Create(setting *domain.PlatformSetting) error {
	query := `
		INSERT INTO platform_settings (setting_key, setting_value, description, data_type, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	setting.CreatedAt = now
	setting.UpdatedAt = now

	err := r.db.QueryRow(
		query,
		setting.SettingKey,
		setting.SettingValue,
		setting.Description,
		setting.DataType,
		setting.IsActive,
		setting.CreatedAt,
		setting.UpdatedAt,
	).Scan(&setting.ID)

	if err != nil {
		return fmt.Errorf("ошибка создания настройки: %w", err)
	}

	return nil
}

func (r *platformSettingsRepository) Update(setting *domain.PlatformSetting) error {
	query := `
		UPDATE platform_settings 
		SET setting_value = $1, description = $2, data_type = $3, is_active = $4, updated_at = $5
		WHERE setting_key = $6
	`

	setting.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		setting.SettingValue,
		setting.Description,
		setting.DataType,
		setting.IsActive,
		setting.UpdatedAt,
		setting.SettingKey,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления настройки: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки результата обновления: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("настройка с ключом '%s' не найдена", setting.SettingKey)
	}

	return nil
}

func (r *platformSettingsRepository) Delete(key string) error {
	query := `DELETE FROM platform_settings WHERE setting_key = $1`

	result, err := r.db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("ошибка удаления настройки: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки результата удаления: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("настройка с ключом '%s' не найдена", key)
	}

	return nil
}
