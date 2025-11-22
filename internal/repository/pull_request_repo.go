package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/queries"
)

var (
	ErrPRNotFound = errors.New("NOT_FOUND")
)

type PullRequestRepo struct {
	db *sql.DB
}

func NewPullRequestRepo(db *sql.DB) *PullRequestRepo {
	return &PullRequestRepo{db: db}
}

// создает запись PR в базе
func (r *PullRequestRepo) CreatePR(pr *domain.PullRequest) error {
	_, err := r.db.Exec(queries.InsertPR, pr.PRID, pr.PRName, pr.AuthorID)
	return err
}

// назначает ревьюверов
func (r *PullRequestRepo) AssignReviewers(prID string, userIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, u := range userIDs {
		if _, err := tx.Exec(queries.InsertPRReviewer, prID, u); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// меняет статус на MERGED
func (r *PullRequestRepo) MergePR(prID string, mergedAt time.Time) error {
	_, err := r.db.Exec(queries.UpdatePRStatusMerged, mergedAt, prID)
	return err
}

// возвращает PR по ID
func (r *PullRequestRepo) GetPRByID(prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	err := r.db.QueryRow(queries.SelectPRByID, prID).Scan(&pr.PRID, &pr.PRName, &pr.AuthorID, &pr.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}
	return &pr, nil
}

// заменяет одного ревьювера на другого
func (r *PullRequestRepo) UpdateReviewer(prID, oldUserID, newUserID string) error {
	_, err := r.db.Exec(queries.UpdatePRReviewer, newUserID, prID, oldUserID)
	return err
}