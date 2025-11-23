package interfaces

import (
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/db"
)

// только чтение
type UserReader interface {
	GetByID(userID string) (*domain.User, error)
	ListByTeam(teamName string) ([]*domain.User, error)
	ListActiveByTeam(teamName string) ([]*domain.User, error)
	GetReviewPR(userID string) ([]*domain.PullRequestShort, error)
}

// только запись
type UserWriter interface {
	Create(exec db.Executor, user *domain.User) error
	SetIsActive(userID string, isActive bool) error
}

// полный интерфейс репо
type UserRepo interface {
	UserReader
	UserWriter
	ListAllUsers() ([]*domain.User, error)
}
