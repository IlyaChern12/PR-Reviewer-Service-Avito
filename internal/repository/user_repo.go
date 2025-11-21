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
	// добавляем пользователя
	// если он уже есть, то обновляем его данные
	_, err := exec.Exec(queries.InsertOrUpdateUser, user.UserID, user.Username, user.TeamName, user.IsActive)

	if err != nil {
		r.logger.Errorf("failed to create or update user %s: %v", user.UserID, err)
		return fmt.Errorf("failed to create or update user: %w", err)
	}
	r.logger.Infof("user %s created/updated successfully", user.UserID)
	return nil
}

// получение юзера по id
func (r *UserRepo) GetByID(userID string) (*domain.User, error){
	var u domain.User

	// запрос на получение
	err := r.db.QueryRow(queries.SelectUserByID, userID).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)

	if err != nil {
		r.logger.Errorf("failed to get user by id %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get user by id %s: %w", userID, err)
	}
	return &u, nil
}

// обновление активности юзера
func (r *UserRepo) SetIsActive(userID string, isActive bool) error {
	// запрос для смены статуса
	_, err := r.db.Exec(queries.UpdateUserIsActive, isActive, userID)
	if err != nil {
		r.logger.Errorf("failed to update isActive for user %s: %v", userID, err)
		return err
	}
	r.logger.Infof("user %s isActive updated to %v", userID, isActive)
	return nil
}

// вспомогательная функция для чтения пользователей
func (r *UserRepo) scanUsers(rows *sql.Rows) ([]*domain.User, error) {
	defer rows.Close()

	users := []*domain.User{}
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
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
		r.logger.Errorf("failed to list users by team %s: %v", teamName, err)
		return nil, err
	}
	return r.scanUsers(rows)
}

// возврат всех активных пользователей команды
func (r *UserRepo) ListActiveByTeam(teamName string) ([]*domain.User, error) {
	rows, err := r.db.Query(queries.SelectActiveUsersByTeam, teamName)
	if err != nil {
		r.logger.Errorf("failed to list active users by team %s: %v", teamName, err)
		return nil, err
	}
	return r.scanUsers(rows)
}