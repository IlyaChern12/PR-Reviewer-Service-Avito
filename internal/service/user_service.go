package service

import (
	"fmt"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/interfaces"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound = fmt.Errorf("user not found")
)

type UserService struct {
    repo   interfaces.UserRepo
    logger *zap.SugaredLogger
}

func NewUserService(repo interfaces.UserRepo, logger *zap.SugaredLogger) *UserService {
    return &UserService{
        repo:   repo,
        logger: logger,
    }
}

// получить пользователя по ID
func (s *UserService) GetByID(userID string) (*domain.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.logger.Warnf("failed to get user %s: %v", userID, err)
		return nil, err
	}
	s.logger.Infof("user %s retrieved", userID)
	return user, nil
}

// обновить статус активности
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

// получение всех пользователей команды
func (s *UserService) ListByTeam(teamName string) ([]*domain.User, error) {
	users, err := s.repo.ListByTeam(teamName)
	if err != nil {
		s.logger.Warnf("failed to list users by team %s: %v", teamName, err)
		return nil, err
	}
	return users, nil
}

// получение всех активных пользователей команды
func (s *UserService) ListActiveByTeam(teamName string) ([]*domain.User, error) {
	users, err := s.repo.ListActiveByTeam(teamName)
	if err != nil {
		s.logger.Warnf("failed to list active users by team %s: %v", teamName, err)
		return nil, err
	}
	return users, nil
}

// получение pr где юзер является ревьюером
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