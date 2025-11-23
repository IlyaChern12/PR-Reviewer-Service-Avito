package repository

import (
	"database/sql"
	"errors"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/queries"
	"go.uber.org/zap"
)

var (
	ErrTeamExists   = errors.New("TEAM_EXISTS")
	ErrTeamNotFound = errors.New("NOT_FOUND")
)


type TeamRepo struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewTeamRepo(db *sql.DB, logger *zap.SugaredLogger) *TeamRepo {
	return &TeamRepo{
		db:     db,
		logger: logger,
	}
}

// Проверка существования команды
func (r *TeamRepo) Exists(teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(queries.SelectTeamExist, teamName).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		r.logger.Errorf("failed to check existence of team %s: %v", teamName, err)
		return false, err
	}
	return exists, nil
}

// создание команды атомарно
func (r *TeamRepo) CreateTeamWithUsers(teamName string, members []*domain.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// вставка команды
	if _, err := tx.Exec(queries.InsertTeam, teamName); err != nil {
		return err
	}

	// вставка/обновление пользователей
	for _, u := range members {
		u.TeamName = teamName
		r.logger.Infof("Inserting user: %s, %s, %s, %v", u.UserID, u.Username, u.TeamName, u.IsActive)
		if _, err := tx.Exec(queries.InsertOrUpdateUser, u.UserID, u.Username, u.TeamName, u.IsActive); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// получение участников команды
func (r *TeamRepo) GetUsersByTeam(teamName string) ([]*domain.User, error) {
	rows, err := r.db.Query(queries.SelectUsersByTeam, teamName)
	if err != nil {
		r.logger.Errorf("failed to query users for team %s: %v", teamName, err)
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			r.logger.Errorf("SQL error: failed to scan user row: %v", err)
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}