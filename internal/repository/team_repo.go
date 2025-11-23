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

// TeamRepo - репо пользователей
type TeamRepo struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

// NewTeamRepo создает новое репо пользователей
func NewTeamRepo(db *sql.DB, logger *zap.SugaredLogger) *TeamRepo {
	return &TeamRepo{
		db:     db,
		logger: logger,
	}
}

// Exists запрашивает проверку на существование
func (r *TeamRepo) Exists(teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(queries.SelectTeamExist, teamName).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		r.logger.Errorf("failed to check existence of team %s: %v", teamName, err)
		return false, err
	}
	return exists, nil
}

// CreateTeamWithUsers создает команды атомарно
func (r *TeamRepo) CreateTeamWithUsers(teamName string, members []*domain.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			r.logger.Errorf("rollback failed: %v", err)
		}
	}()

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

// GetUsersByTeam получение участников команды
func (r *TeamRepo) GetUsersByTeam(teamName string) ([]*domain.User, error) {
	rows, err := r.db.Query(queries.SelectUsersByTeam, teamName)
	if err != nil {
		r.logger.Errorf("failed to query users for team %s: %v", teamName, err)
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Errorf("rows close failed: %v", err)
		}
	}()

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

// ListAllTeams получает все созданные команды
func (r *TeamRepo) ListAllTeams() ([]*domain.Team, error) {
	rows, err := r.db.Query(queries.SelectAllTeams)
	if err != nil {
		r.logger.Errorf("failed to list all teams: %v", err)
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Errorf("rows close failed: %v", err)
		}
	}()

	var teams []*domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.TeamName); err != nil {
			r.logger.Errorf("failed to scan team row: %v", err)
			return nil, err
		}
		teams = append(teams, &t)
	}
	return teams, nil
}
