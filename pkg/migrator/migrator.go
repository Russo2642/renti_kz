package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/russo2642/renti_kz/internal/config"
)

func RunMigrations(dbConfig config.DatabaseConfig, migrationConfig config.MigrationConfig) error {
	if !migrationConfig.AutoMigrate {
		log.Println("Автоматическое применение миграций отключено в настройках")
		return nil
	}

	log.Println("Начало выполнения миграций...")

	dsn := dbConfig.DSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("ошибка при подключении к базе данных для миграций: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ошибка при проверке соединения с базой данных: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка при создании драйвера миграций: %w", err)
	}

	migrationsPath := fmt.Sprintf("file://%s", migrationConfig.MigrationsPath)

	if _, err := os.Stat(migrationConfig.MigrationsPath); os.IsNotExist(err) {
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("не удалось определить путь к исполняемому файлу: %w", err)
		}

		execDir := filepath.Dir(execPath)
		possiblePath := filepath.Join(execDir, migrationConfig.MigrationsPath)

		if _, err := os.Stat(possiblePath); !os.IsNotExist(err) {
			migrationsPath = fmt.Sprintf("file://%s", possiblePath)
		} else {
			return fmt.Errorf("директория с миграциями не найдена: %s", migrationConfig.MigrationsPath)
		}
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("ошибка при создании экземпляра мигратора: %w", err)
	}

	m.LockTimeout = 60 * time.Second

	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("ошибка при выполнении миграций: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("Миграции не требуются, база данных в актуальном состоянии")
	} else {
		log.Println("Миграции успешно выполнены")
	}

	return nil
}
