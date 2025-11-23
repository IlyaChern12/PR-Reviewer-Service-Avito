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

// UserRepo репозиторий для работы с пользователями
type UserRepo struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

// NewUserRepo создает новый UserRepo
func NewUserRepo(db *sql.DB, logger *zap.SugaredLogger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logger,
	}
}

// Create создает нового пользователя или обновляет существующего
func (r *UserRepo) Create(exec db.Executor, user *domain.User) error {
	if _, err := exec.Exec(queries.InsertOrUpdateUser, user.UserID, user.Username, user.TeamName, user.IsActive); err != nil {
		r.logger.Errorf("SQL error: failed to create/update user %s: %v", user.UserID, err)
		return fmt.Errorf("failed to create/update user: %w", err)
	}
	return nil
}

// GetByID получает пользователя по ID.
func (r *UserRepo) GetByID(userID string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(queries.SelectUserByID, userID).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Errorf("SQL error: failed to get user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

// SetIsActive обновляет статус активности пользователя.
func (r *UserRepo) SetIsActive(userID string, isActive bool) error {
	res, err := r.db.Exec(queries.UpdateUserIsActive, isActive, userID)
	if err != nil {
		r.logger.Errorf("SQL error: failed to update isActive for user %s: %v", userID, err)
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// SetIsActiveByTeam массово обновляет статус для всех пользователей команды
func (r *UserRepo) SetIsActiveByTeam(teamName string, isActive bool) error {
	_, err := r.db.Exec(queries.SetIsActiveStatusByTeamName, isActive, teamName)
	return err
}

// scanUsers является вспомогательной функцией для чтения пользователей из sql.Rows.
func (r *UserRepo) scanUsers(rows *sql.Rows) ([]*domain.User, error) {
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

// ListByTeam возвращает всех пользователей команды
func (r *UserRepo) ListByTeam(teamName string) ([]*domain.User, error) {
	rows, err := r.db.Query(queries.SelectUsersByTeam, teamName)
	if err != nil {
		r.logger.Errorf("SQL error: failed to list users by team %s: %v", teamName, err)
		return nil, err
	}
	return r.scanUsers(rows)
}

// ListActiveByTeam возвращает всех активных пользователей команды
func (r *UserRepo) ListActiveByTeam(teamName string) ([]*domain.User, error) {
	rows, err := r.db.Query(queries.SelectActiveUsersByTeam, teamName)
	if err != nil {
		r.logger.Errorf("SQL error: failed to list active users by team %s: %v", teamName, err)
		return nil, err
	}
	return r.scanUsers(rows)
}

// GetReviewPR получает список PR, где пользователь является ревьюером
func (r *UserRepo) GetReviewPR(userID string) ([]*domain.PullRequestShort, error) {
	rows, err := r.db.Query(queries.SelectReviewPRsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Errorf("rows close failed: %v", err)
		}
	}()

	var prs []*domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.PRID, &pr.PRName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, &pr)
	}
	return prs, nil
}

// ListAllUsers возвращает список всех пользователей.
func (r *UserRepo) ListAllUsers() ([]*domain.User, error) {
	rows, err := r.db.Query(`SELECT user_id, username, team_name, is_active FROM users`)
	if err != nil {
		r.logger.Errorf("failed to list all users: %v", err)
		return nil, err
	}
	return r.scanUsers(rows)
}
