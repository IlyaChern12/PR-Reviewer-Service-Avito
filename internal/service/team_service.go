package service

import (
	"database/sql"
	"errors"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/interfaces"
	"go.uber.org/zap"
)

var ErrTeamExists = errors.New("TEAM_EXISTS")
var ErrTeamNotFound = errors.New("NOT_FOUND")

// TeamService - сервис пользователей
type TeamService struct {
	repo   interfaces.TeamRepo
	db     *sql.DB
	logger *zap.SugaredLogger
}

// NewTeamService cоздает новый сервис пользователей
func NewTeamService(repo interfaces.TeamRepo, db *sql.DB, logger *zap.SugaredLogger) *TeamService {
	return &TeamService{repo: repo, db: db, logger: logger}
}

// CreateTeam создает команду
func (s *TeamService) CreateTeam(team *domain.Team, members []*domain.User) error {
	exists, err := s.repo.Exists(team.TeamName)
	if err != nil {
		return err
	}
	if exists {
		return ErrTeamExists
	}

	return s.repo.CreateTeamWithUsers(team.TeamName, members)
}

// TeamExists проверяет существование команды
func (s *TeamService) TeamExists(teamName string) (bool, error) {
	// читаем из репозитория
	exists, err := s.repo.Exists(teamName)
	if err != nil {
		// логируем на уровне сервиса
		s.logger.Errorf("failed to check if team exists %s: %v", teamName, err)
		return false, err
	}
	return exists, nil
}

// GetTeam получает команду по имени
func (s *TeamService) GetTeam(teamName string) (*domain.Team, error) {
	members, err := s.repo.GetUsersByTeam(teamName)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, ErrTeamNotFound
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

// GetStats получает статистику по созданным командам
func (s *TeamService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	teams, err := s.repo.ListAllTeams()
	if err != nil {
		return nil, err
	}

	total := len(teams)
	usersPerTeam := make(map[string]int)
	for _, t := range teams {
		users, err := s.repo.GetUsersByTeam(t.TeamName)
		if err != nil {
			return nil, err
		}
		usersPerTeam[t.TeamName] = len(users)
	}

	stats["total_teams"] = total
	stats["users_per_team"] = usersPerTeam

	return stats, nil
}
