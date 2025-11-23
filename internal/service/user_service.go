package service

import (
	"fmt"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/interfaces"
	"go.uber.org/zap"
)

// ErrUserNotFound возвращается, если пользователь не найден
var ErrUserNotFound = fmt.Errorf("user not found")

// UserService для работы с пользователями
type UserService struct {
	repo   interfaces.UserRepo
	logger *zap.SugaredLogger
}

// NewUserService создаёт новый сервис пользователей
func NewUserService(repo interfaces.UserRepo, logger *zap.SugaredLogger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

// GetByID возвращает пользователя по его ID
func (s *UserService) GetByID(userID string) (*domain.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.logger.Warnf("failed to get user %s: %v", userID, err)
		return nil, err
	}
	s.logger.Infof("user %s retrieved", userID)
	return user, nil
}

// SetIsActive обновляет статус активности пользователя
func (s *UserService) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	if err := s.repo.SetIsActive(userID, isActive); err != nil {
		s.logger.Warnf("failed to set isActive for user %s: %v", userID, err)
		return nil, err
	}
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.logger.Warnf("failed to get updated user %s: %v", userID, err)
		return nil, err
	}
	s.logger.Infof("user %s isActive updated to %v", userID, isActive)
	return user, nil
}

// DeactivateTeam массово деактивирует всех пользователей команды
func (s *UserService) DeactivateTeam(teamName string) error {
    return s.repo.SetIsActiveByTeam(teamName, false)
}

// ListByTeam возвращает всех пользователей команды
func (s *UserService) ListByTeam(teamName string) ([]*domain.User, error) {
	users, err := s.repo.ListByTeam(teamName)
	if err != nil {
		s.logger.Warnf("failed to list users by team %s: %v", teamName, err)
		return nil, err
	}
	return users, nil
}

// ListActiveByTeam возвращает всех активных пользователей команды
func (s *UserService) ListActiveByTeam(teamName string) ([]*domain.User, error) {
	users, err := s.repo.ListActiveByTeam(teamName)
	if err != nil {
		s.logger.Warnf("failed to list active users by team %s: %v", teamName, err)
		return nil, err
	}
	return users, nil
}

// ListReviewPR возвращает PR, где пользователь является ревьюером
func (s *UserService) ListReviewPR(userID string) ([]*domain.PullRequestShort, error) {
	// сначала проверяем, что пользователь существует
	_, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	prs, err := s.repo.GetReviewPR(userID)
	if err != nil {
		s.logger.Warnf("failed to get review PR for user %s: %v", userID, err)
		return nil, err
	}

	return prs, nil
}

// GetReviewPR возвращает PR, где пользователь является ревьюером (обёртка)
func (s *UserService) GetReviewPR(userID string) ([]*domain.PullRequestShort, error) {
	return s.ListReviewPR(userID)
}

// GetStats возвращает статистику по пользователям
func (s *UserService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, err
	}

	total := len(users)
	activeCount := 0
	inactiveCount := 0
	for _, u := range users {
		if u.IsActive {
			activeCount++
		} else {
			inactiveCount++
		}
	}

	stats["total_users"] = total
	stats["active_users"] = activeCount
	stats["inactive_users"] = inactiveCount

	return stats, nil
}
