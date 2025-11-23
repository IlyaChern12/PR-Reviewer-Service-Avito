package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/queries"
	"go.uber.org/zap"
)

var (
	ErrPRExists   = errors.New("PR_EXISTS")
	ErrPRNotFound = errors.New("PR_NOT_FOUND")
)

// PullRequestRepo - репо PR
type PullRequestRepo struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

// NewPullRequestRepo создает новый репо PR
func NewPullRequestRepo(db *sql.DB) *PullRequestRepo {
	return &PullRequestRepo{db: db}
}

// CreatePR создает запись PR в базе
func (r *PullRequestRepo) CreatePR(pr *domain.PullRequest) error {
	assignedIDs := []string{}
	for _, u := range pr.AssignReviewers {
		assignedIDs = append(assignedIDs, u.UserID)
	}

	var exists bool
	if err := r.db.QueryRow(queries.SelectPRExist, pr.PRID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrPRExists
	}

	_, err := r.db.Exec(queries.InsertPR, pr.PRID, pr.PRName, pr.AuthorID)
	if err != nil {
		return err
	}

	for _, id := range assignedIDs {
		if _, err := r.db.Exec(queries.InsertPRReviewer, pr.PRID, id); err != nil {
			return err
		}
	}

	return err
}

// AssignReviewers назначает ревьюверов
func (r *PullRequestRepo) AssignReviewers(prID string, userIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			r.logger.Errorf("rollback failed: %v", err)
		}
	}()

	for _, u := range userIDs {
		if _, err := tx.Exec(queries.InsertPRReviewer, prID, u); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// MergePR меняет статус на MERGED
func (r *PullRequestRepo) MergePR(prID string, mergedAt time.Time) error {
	_, err := r.db.Exec(queries.UpdatePRStatusMerged, mergedAt, prID)
	return err
}

// GetPRByID возвращает PR по ID
func (r *PullRequestRepo) GetPRByID(prID string) (*domain.PullRequest, error) {
	pr := &domain.PullRequest{}
	var mergedAt sql.NullTime

	err := r.db.QueryRow(queries.SelectPRByID, prID).
		Scan(&pr.PRID, &pr.PRName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	// выбираем назначенных ревьюверов
	rows, err := r.db.Query(queries.SelectPRReviewersFull, prID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Errorf("rows close failed: %v", err)
		}
	}()

	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		pr.AssignReviewers = append(pr.AssignReviewers, u)
	}

	return pr, nil
}

// UpdateReviewer заменяет одного ревьювера на другого
func (r *PullRequestRepo) UpdateReviewer(prID, oldUserID, newUserID string) error {
	_, err := r.db.Exec(queries.UpdatePRReviewer, newUserID, prID, oldUserID)
	return err
}

// ListOpenPRsByTeam возвращает все PR со статусом OPEN для авторов команды
func (r *PullRequestRepo) ListOpenPRsByTeam(teamName string) ([]*domain.PullRequest, error) {
	rows, err := r.db.Query(queries.GetOpenPRsByTeamName, teamName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Errorf("rows close failed: %v", err)
		}
	}()

	prMap := make(map[string]*domain.PullRequest)
	for rows.Next() {
		var prID, prName, authorID, status string
		var reviewer domain.User
		var reviewerID sql.NullString

		if err := rows.Scan(&prID, &prName, &authorID, &status, &reviewerID, &reviewer.Username, &reviewer.TeamName, &reviewer.IsActive); err != nil {
			return nil, err
		}

		pr, exists := prMap[prID]
		if !exists {
			pr = &domain.PullRequest{
				PRID:            prID,
				PRName:          prName,
				AuthorID:        authorID,
				Status:          status,
				AssignReviewers: []*domain.User{},
			}
			prMap[prID] = pr
		}

		if reviewerID.Valid {
			reviewer.UserID = reviewerID.String
			pr.AssignReviewers = append(pr.AssignReviewers, &reviewer)
		}
	}

	prs := make([]*domain.PullRequest, 0, len(prMap))
	for _, pr := range prMap {
		prs = append(prs, pr)
	}
	return prs, nil
}

// ListAllPRs вывод всех пулл реквестов
func (r *PullRequestRepo) ListAllPRs() ([]*domain.PullRequest, error) {
	rows, err := r.db.Query(queries.SelectAllRPs)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Errorf("rows close failed: %v", err)
		}
	}()

	var prs []*domain.PullRequest
	for rows.Next() {
		var prID string
		if err := rows.Scan(&prID); err != nil {
			return nil, err
		}
		pr, err := r.GetPRByID(prID)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, nil
}
