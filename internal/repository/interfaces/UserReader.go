package interfaces

import "github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"

// специальный интерфейс для упрощения зависимостей
// между UserRepo и PullRequestRepo
type UserReader interface {
	ListActiveByTeam(teamName string) ([]*domain.User, error)
}