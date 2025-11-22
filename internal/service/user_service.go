package service

import (
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/interfaces"
	"go.uber.org/zap"
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
		s.logger.Warnf("Failed to get user %s: %v", userID, err)
		return nil, err
	}
	s.logger.Infof("User %s retrieved", userID)
	return user, nil
}

// обновить статус активности
func (s *UserService) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	if err := s.repo.SetIsActive(userID, isActive); err != nil {
		s.logger.Warnf("Failed to set isActive for user %s: %v", userID, err)
		return nil, err
	}
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.logger.Warnf("Failed to get updated user %s: %v", userID, err)
		return nil, err
	}
	s.logger.Infof("User %s isActive updated to %v", userID, isActive)
	return user, nil
}

// получение всех пользователей команды
func (s *UserService) ListByTeam(teamName string) ([]*domain.User, error) {
	users, err := s.repo.ListByTeam(teamName)
	if err != nil {
		s.logger.Warnf("Failed to list users by team %s: %v", teamName, err)
		return nil, err
	}
	return users, nil
}

// получение всех активных пользователей команды
func (s *UserService) ListActiveByTeam(teamName string) ([]*domain.User, error) {
	users, err := s.repo.ListActiveByTeam(teamName)
	if err != nil {
		s.logger.Warnf("Failed to list active users by team %s: %v", teamName, err)
		return nil, err
	}
	return users, nil
}