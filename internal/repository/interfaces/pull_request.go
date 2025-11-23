package interfaces

import (
	"time"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
)

type PullRequestReader interface {
    GetPRByID(prID string) (*domain.PullRequest, error)
}

type PullRequestWriter interface {
    CreatePR(pr *domain.PullRequest) error
    MergePR(prID string, mergedAt time.Time) error
    AssignReviewers(prID string, userIDs []string) error
	UpdateReviewer(prID, oldUserID, newUserID string) error
}

type PullRequestRepo interface {
    PullRequestReader
    PullRequestWriter
    ListAllPRs() ([]*domain.PullRequest, error)
}