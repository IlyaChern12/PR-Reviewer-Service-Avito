package integration

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

// тестовая бд
var testDB *sql.DB

// InitTestDB подключает базу для тестов и применение миграций
func InitTestDB() *sql.DB {
	// получаем cтроку подключения
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("TEST_DATABASE_URL not set")
	}

	// открываем базу
	var err error
	testDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to test DB: %v", err)
	}

	// поднимаем миграции
	if err := applyMigrations(); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	return testDB
}

// applyMigrations применяет все миграции
func applyMigrations() error {
	// создаем адаптер
	driver, err := postgres.WithInstance(testDB, &postgres.Config{})
	if err != nil {
		return err
	}

	// объект миграций
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	// поднимаем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// ResetDB очищает все таблицы перед каждым тестом
func ResetDB() {
	// проверка подлкючения
	if testDB == nil {
		log.Fatal("testDB not initialized")
	}

	// исполняем запросы на очистку
	_, err := testDB.Exec(`
        TRUNCATE TABLE pull_requests RESTART IDENTITY CASCADE;
        TRUNCATE TABLE users RESTART IDENTITY CASCADE;
        TRUNCATE TABLE teams RESTART IDENTITY CASCADE;
		TRUNCATE TABLE pull_request_reviewers RESTART IDENTITY CASCADE;
    `)
	if err != nil {
		log.Fatalf("failed to truncate tables: %v", err)
	}
}
