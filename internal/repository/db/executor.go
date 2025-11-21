package db

import "database/sql"

// всопомогательный интерфейс для атомарности
type Executor interface {
    Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
}