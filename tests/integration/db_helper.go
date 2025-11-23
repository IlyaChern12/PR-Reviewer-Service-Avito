package integration

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

var testDB *sql.DB

// InitTestDB подключает базу для тестов и применение миграций
func InitTestDB() {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("TEST_DATABASE_URL not set")
	}

	var err error
	testDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to test DB: %v", err)
	}

	if err := applyMigrations(dbURL); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}
}

// применяем все миграции
func applyMigrations(dbURL string) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations", // путь к миграциям
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// ResetDB очищает все таблицы перед каждым тестом
func ResetDB() {
	if testDB == nil {
		log.Fatal("testDB not initialized")
	}

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
