package service

import (
	"errors"
	"time"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/interfaces"
	"go.uber.org/zap"
)

var (
	ErrPRNotFound      = errors.New("PR_NOT_FOUND")
	ErrAuthorNotFound  = errors.New("AUTHOR_NOT_FOUND")
	ErrPRExists        = errors.New("PR_EXISTS")
	ErrPRMerged        = errors.New("PR_MERGED")
	ErrNoCandidate     = errors.New("NO_CANDIDATE")
)

type PullRequestService struct {
	prRepo interfaces.PullRequestRepo // репо PR
	userRepo interfaces.UserReader    // для получения активных ревьюверов
	logger  *zap.SugaredLogger
}

func NewPullRequestService(prRepo interfaces.PullRequestRepo, userRepo interfaces.UserReader, logger *zap.SugaredLogger) *PullRequestService {
	return &PullRequestService{
		prRepo: prRepo,
		userRepo: userRepo,
		logger:  logger,
	}
}

// cоздание PR с назначением до 2 ревьюверов
func (s *PullRequestService) CreatePR(pr *domain.PullRequest) error {
	// получаем активных участников
	users, err := s.userRepo.ListActiveByTeam(pr.AuthorID)
	if err != nil {
		return err
	}

	// исключаем автора
	var reviewers []string
	for _, u := range users {
		if u.UserID != pr.AuthorID && len(reviewers) < 2 {
			reviewers = append(reviewers, u.UserID)
		}
	}

	// создаем PR
	if err := s.prRepo.CreatePR(pr); err != nil {
		return err
	}

	// назначаем ревьюверов
	if len(reviewers) > 0 {
		if err := s.prRepo.AssignReviewers(pr.PRID, reviewers); err != nil {
			return err
		}
	}

	return nil
}

// cлияние
func (s *PullRequestService) MergePR(prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetPRByID(prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	now := time.Now()
	if err := s.prRepo.MergePR(prID, now); err != nil {
		return nil, err
	}

	pr.Status = "MERGED"
	pr.MergedAt = &now
	return pr, nil
}

// замена ревьювера на активного из его команды
func (s *PullRequestService) ReassignReviewer(prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetPRByID(prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == "MERGED" {
		return nil, "", errors.New("PR_MERGED")
	}

	// получаем активных участников
	users, err := s.userRepo.ListActiveByTeam(oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	var newReviewerID string
	for _, u := range users {
		if u.UserID != oldReviewerID {
			newReviewerID = u.UserID
			break
		}
	}
	if newReviewerID == "" {
		return nil, "", errors.New("NO_CANDIDATE")
	}

	// обновляем запись
	if err := s.prRepo.UpdateReviewer(prID, oldReviewerID, newReviewerID); err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}