package service

import (
	"errors"
	"time"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/interfaces"
	"go.uber.org/zap"
)

var (
	ErrPRNotFound     = errors.New("PR_NOT_FOUND")
	ErrAuthorNotFound = errors.New("AUTHOR_NOT_FOUND")
	ErrPRExists       = errors.New("PR_EXISTS")
	ErrPRMerged       = errors.New("PR_MERGED")
	ErrNoCandidate    = errors.New("NO_CANDIDATE")
	ErrNotAssigned    = errors.New("reviewer is not assigned to PR")
)

type PullRequestService struct {
	prRepo   interfaces.PullRequestRepo // репо PR
	userRepo interfaces.UserReader      // для получения активных ревьюверов
	logger   *zap.SugaredLogger
}

func NewPullRequestService(prRepo interfaces.PullRequestRepo, userRepo interfaces.UserReader, logger *zap.SugaredLogger) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

// cоздание PR с назначением до 2 ревьюверов
func (s *PullRequestService) CreatePR(pr *domain.PullRequest) error {
	existing, err := s.prRepo.GetPRByID(pr.PRID)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
		} else {
			return err
		}
	} else if existing != nil {
		return ErrPRExists
	}

	author, err := s.userRepo.GetByID(pr.AuthorID)
	if err != nil {
		return ErrAuthorNotFound
	}

	// получаем активных участников
	users, err := s.userRepo.ListActiveByTeam(author.TeamName)
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
	pr.Status = "OPEN"
	if err := s.prRepo.CreatePR(pr); err != nil {
		return err
	}

	// назначаем ревьюверов
	if len(reviewers) > 0 {
		if err := s.prRepo.AssignReviewers(pr.PRID, reviewers); err != nil {
			return err
		}

		// заполняем поле AssignReviewers для ответа
		pr.AssignReviewers = []*domain.User{}
		for _, u := range users {
			for _, id := range reviewers {
				if u.UserID == id {
					pr.AssignReviewers = append(pr.AssignReviewers, u)
					break
				}
			}
		}
	} else {
		pr.AssignReviewers = []*domain.User{}
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
		return nil, "", ErrPRMerged
	}

	// получаем активных участников
	oldUser, err := s.userRepo.GetByID(oldReviewerID)
	if err != nil {
		return nil, "", ErrNotAssigned
	}

	users, err := s.userRepo.ListActiveByTeam(oldUser.TeamName)
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

// сервисный метод для статистики
func (s *PullRequestService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 1. Количество назначений на PR по пользователям
	assignments := make(map[string]int)
	prs, err := s.prRepo.ListAllPRs()
	if err != nil {
		return nil, err
	}
	for _, pr := range prs {
		for _, u := range pr.AssignReviewers {
			assignments[u.UserID]++
		}
	}
	stats["assignments_per_user"] = assignments

	// 2. Количество PR по статусу
	statusCount := map[string]int{
		"OPEN":   0,
		"MERGED": 0,
	}
	for _, pr := range prs {
		statusCount[pr.Status]++
	}
	stats["prs_per_status"] = statusCount

	stats["total_pull_requests"] = len(prs)

	return stats, nil
}
