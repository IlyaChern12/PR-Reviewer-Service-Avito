package db

import (
	"database/sql"
	"fmt"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/config"
	_ "github.com/lib/pq"
)

func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	// строка подключения к бд
	dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName,
    )
	return sql.Open("postgres", dsn)
}