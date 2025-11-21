package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/db"
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

// проверка существования команды
func (r *TeamRepo) Exists(exec db.Executor, teamName string) (bool, error) {
	var exists bool
	err := exec.QueryRow(queries.SelectTeamExist, teamName).Scan(&exists)
	if err != nil {
		r.logger.Errorf("failed to check existence of team %s: %v", teamName, err)
		return false, err
	}
	return exists, nil
}

// создание команды с участниками
func (r *TeamRepo) Create(team *domain.Team, members []*domain.User) error {
	// транзакция для атомарности
	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Errorf("failed to begin transaction for creating team %s: %v", team.TeamName, err)
		return err
	}
	defer func() { 		// откат при ошибке
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			r.logger.Warnf("rollback failed for team %s: %v", team.TeamName, err)
		}
	}()

	// проверяем, что команда не существует
	exists, err := r.Exists(tx, team.TeamName)
	if err != nil {
		return err
	}
	if exists {
		return ErrTeamExists
	}

	// создаём команду
	if _, err := tx.Exec(queries.InsertTeam, team.TeamName); err != nil {
		r.logger.Errorf("failed to insert team %s: %v", team.TeamName, err)
		return fmt.Errorf("failed to create team: %w", err)
	}

	// добавляем участников
	for _, u := range members {
		if _, err := tx.Exec(queries.InsertOrUpdateUser, u.UserID, u.Username, u.TeamName, u.IsActive); err != nil {
			r.logger.Errorf("failed to insert/update user %s in team %s: %v", u.UserID, team.TeamName, err)
			return fmt.Errorf("failed to create or update user: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Errorf("failed to commit transaction for team %s: %v", team.TeamName, err)
		return err
	}

	r.logger.Infof("team %s created successfully with %d members", team.TeamName, len(members))
	return nil
}

// возвращает команду с участниками
func (r *TeamRepo) GetByName(teamName string) (*domain.Team, error) {
	exists, err := r.Exists(r.db, teamName)
	if err != nil {
		return nil, err
	}
	// если команда не существует выводим ошибку
	if !exists {
		return nil, ErrTeamNotFound
	}

	// получаем участников
	rows, err := r.db.Query(queries.SelectUsersByTeam,teamName)
	if err != nil {
		r.logger.Errorf("failed to query users for team %s: %v", teamName, err)
		return nil, err
	}
	defer rows.Close() // закрываем ресурс

	members, err := scanUsers(rows, teamName)
	if err != nil {
		r.logger.Errorf("failed to scan users for team %s: %v", teamName, err)
		return nil, err
	}

	return &domain.Team{
		TeamName: teamName,
		Members: members,
	}, nil
}

// читает строки пользователей в срез
func scanUsers(rows *sql.Rows, teamName string) ([]*domain.User, error) {
	var users []*domain.User
	for rows.Next() {
		u := &domain.User{TeamName: teamName}
		if err := rows.Scan(&u.UserID, &u.Username, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}