package interfaces

import (
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
)

// только запись
type TeamWriter interface {
	CreateTeamWithUsers(teamName string, members []*domain.User) error
}

// чтение
type TeamReader interface {
	Exists(teamName string) (bool, error)
	GetUsersByTeam(teamName string) ([]*domain.User, error)
}

// полный интерфейс репо
type TeamRepo interface {
	TeamReader
	TeamWriter
	ListAllTeams() ([]*domain.Team, error)
}
