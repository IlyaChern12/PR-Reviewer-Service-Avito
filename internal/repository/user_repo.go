package repository

import (
	"database/sql"
	"fmt"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/db"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/queries"
	"go.uber.org/zap"
)

type UserRepo struct {
	db *sql.DB
	logger *zap.SugaredLogger
}

func NewUserRepo(db *sql.DB, logger *zap.SugaredLogger) *UserRepo {
	return &UserRepo{
		db: db,
		logger: logger,
	}
}

// создаем нового пользователя
func (r *UserRepo) Create(exec db.Executor, user *domain.User) error {
	if _, err := exec.Exec(queries.InsertOrUpdateUser, user.UserID, user.Username, user.TeamName, user.IsActive); err != nil {
		r.logger.Errorf("SQL error: failed to create/update user %s: %v", user.UserID, err)
		return fmt.Errorf("failed to create/update user: %w", err)
	}
	return nil
}

// получение юзера по id
func (r *UserRepo) GetByID(userID string) (*domain.User, error) {
	var u domain.User
	if err := r.db.QueryRow(queries.SelectUserByID, userID).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		r.logger.Errorf("SQL error: failed to get user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

// обновление активности юзера
func (r *UserRepo) SetIsActive(userID string, isActive bool) error {
	if _, err := r.db.Exec(queries.UpdateUserIsActive, isActive, userID); err != nil {
		r.logger.Errorf("SQL error: failed to update isActive for user %s: %v", userID, err)
		return err
	}
	return nil
}

// вспомогательная функция для чтения пользователей
func (r *UserRepo) scanUsers(rows *sql.Rows) ([]*domain.User, error) {
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

// возврат всех пользователей команды
func (r *UserRepo) ListByTeam(teamName string) ([]*domain.User, error) {
	rows, err := r.db.Query(queries.SelectUsersByTeam, teamName)
	if err != nil {
		r.logger.Errorf("SQL error: failed to list users by team %s: %v", teamName, err)
		return nil, err
	}
	return r.scanUsers(rows)
}

// возврат всех активных пользователей команды
func (r *UserRepo) ListActiveByTeam(teamName string) ([]*domain.User, error) {
	rows, err := r.db.Query(queries.SelectActiveUsersByTeam, teamName)
	if err != nil {
		r.logger.Errorf("SQL error: failed to list active users by team %s: %v", teamName, err)
		return nil, err
	}
	return r.scanUsers(rows)
}