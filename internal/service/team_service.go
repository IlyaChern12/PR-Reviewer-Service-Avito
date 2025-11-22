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

type TeamService struct {
    repo   interfaces.TeamRepo
    db     *sql.DB
    logger *zap.SugaredLogger
}

func NewTeamService(repo interfaces.TeamRepo, db *sql.DB, logger *zap.SugaredLogger) *TeamService {
    return &TeamService{repo: repo, db: db, logger: logger}
}

// создание команды
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

// получение команды
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