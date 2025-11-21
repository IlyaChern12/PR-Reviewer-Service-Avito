package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/db"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/interfaces"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/queries"
	"go.uber.org/zap"
)

var (
	ErrPRExists           = errors.New("PR_EXISTS")
	ErrPRNotFound         = errors.New("NOT_FOUND")
	ErrPRMerged           = errors.New("PR_MERGED")
	ErrReviewerNotAssigned = errors.New("NOT_ASSIGNED")
	ErrNoCandidate        = errors.New("NO_CANDIDATE")
)

type PullRequestRepo struct {
	db       *sql.DB
	userRepo interfaces.UserReader // для получения активных ревьюверов
	logger   *zap.SugaredLogger
}

func NewPullRequestRepo(db *sql.DB, ur interfaces.UserReader, logger *zap.SugaredLogger) *PullRequestRepo {
	return &PullRequestRepo{
		db:       db,
		userRepo: ur,
		logger:   logger,
	}
}


// проверка существования PR
func (r *PullRequestRepo) PRExists(exec db.Executor, prID string) (bool, error) {
	var exists bool
	if err := exec.QueryRow(queries.SelectPRExist, prID).Scan(&exists); err != nil {
		r.logger.Errorf("failed to check exist of PR %s: %v", prID, err)
		return false, err
	}

	return exists, nil
}

// создаёт новый PR и назначает до 2 активных ревьюверов
func (r *PullRequestRepo) CreatePR(pr *domain.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Errorf("failed to begin transaction for PR %s: %v", pr.PRID, err)
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			r.logger.Warnf("rollback failed for PR %s: %v", pr.PRID, err)
		}
	}()

	// проверка что PR не существует
	exists, err := r.PRExists(tx, pr.PRID)
	if err != nil {
		return err
	}
	if exists {
		return ErrPRExists
	}

	// создаём PR
	if _, err := tx.Exec(queries.InsertPR, pr.PRID, pr.PRName, pr.AuthorID); err != nil {
		r.logger.Errorf("failed to insert PR %s: %v", pr.PRID, err)
		return fmt.Errorf("failed to create PR: %w", err)
	}

	// получаем активных участников
	reviewers, err := r.getReviewersWithoutAuthor(pr.AuthorID)
	if err != nil {
		return err
	}

	// исключаем автора
	if err := r.assignReviewers(tx, pr.PRID, reviewers, 2); err != nil {
		r.logger.Errorf("failed to assign reviewers for PR %s: %v", pr.PRID, err)
		return err
	}

	if err := tx.Commit(); err != nil {
		r.logger.Errorf("failed to commit transaction for PR %s: %v", pr.PRID, err)
		return err
	}

	r.logger.Infof("PR %s created and reviewers assigned", pr.PRID)
	return nil
}

// возвращает активных участников команды без автора
func (r *PullRequestRepo) getReviewersWithoutAuthor(authorID string) ([]*domain.User, error) {
	users, err := r.userRepo.ListActiveByTeam(authorID)
	if err != nil {
		return nil, err
	}

	var reviewers []*domain.User
	for _, u := range users {
		if u.UserID != authorID {
			reviewers = append(reviewers, u)
		}
	}
	return reviewers, nil
}

// назначает n ревьюверов на PR
func (r *PullRequestRepo) assignReviewers(exec db.Executor, prID string, reviewers []*domain.User, n int) error {
	for i, u := range reviewers {
		if i >= n {
			break
		}
		if _, err := exec.Exec(queries.InsertPRReviewer, prID, u.UserID); err != nil {
			return fmt.Errorf("failed to assign reviewer: %w", err)
		}
	}
	return nil
}

// помечает PR как MERGED идемпотентно
func (r *PullRequestRepo) MergePR(prID string) (*domain.PullRequest, error) {
	pr, err := r.getPRByID(r.db, prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	// фиксируем время слияния
	mergedAt := time.Now()
	if _, err := r.db.Exec(queries.UpdatePRStatusMerged, mergedAt, prID); err != nil {
		r.logger.Errorf("failed to merge PR %s: %v", prID, err)
		return nil, err
	}

	// меняем состояние
	pr.Status = "MERGED"
	pr.MergedAt = &mergedAt
	r.logger.Infof("PR %s merged", prID)
	return pr, nil
}

// возвращает PR по ID
func (r *PullRequestRepo) getPRByID(exec db.Executor, prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	err := exec.QueryRow(queries.SelectPRByID, prID).Scan(&pr.PRID, &pr.PRName, &pr.AuthorID, &pr.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}
	return &pr, nil
}

// переназначает ревьювера на другого из его команды
func (r *PullRequestRepo) ReassignReviewer(prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Errorf("failed to begin transaction for reassign PR %s: %v", prID, err)
		return nil, "", err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			r.logger.Warnf("rollback failed for PR %s: %v", prID, err)
		}
	}()

	pr, err := r.getPRByID(tx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == "MERGED" {
		return nil, "", ErrPRMerged
	}

	if assigned, err := r.isReviewerAssigned(tx, prID, oldReviewerID); err != nil {
		return nil, "", err
	} else if !assigned {
		return nil, "", ErrReviewerNotAssigned
	}

	// выбираем нового ревьюера
	newCandidate, err := r.getNewReviewer(oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	// записываем в БД
	if _, err := tx.Exec(queries.UpdatePRReviewer, newCandidate.UserID, prID, oldReviewerID); err != nil {
		r.logger.Errorf("failed to update reviewer for PR %s: %v", prID, err)
		return nil, "", err
	}

	if err := tx.Commit(); err != nil {
		r.logger.Errorf("failed to commit reassignment for PR %s: %v", prID, err)
		return nil, "", err
	}

	r.logger.Infof("PR %s reviewer reassigned from %s to %s", prID, oldReviewerID, newCandidate.UserID)
	return pr, newCandidate.UserID, nil
}

// проверка что пользователь - ревьюер
func (r *PullRequestRepo) isReviewerAssigned(exec db.Executor, prID, userID string) (bool, error) {
	var count int
	if err := exec.QueryRow(queries.SelectPRReviewers, prID, userID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// выбор нового ревьюера
func (r *PullRequestRepo) getNewReviewer(oldReviewerID string) (*domain.User, error) {
	candidates, err := r.userRepo.ListActiveByTeam(oldReviewerID)
	if err != nil {
		return nil, err
	}
	// выбираем первого подходящего
	for _, u := range candidates {
		if u.UserID != oldReviewerID {
			return u, nil
		}
	}
	return nil, ErrNoCandidate
}