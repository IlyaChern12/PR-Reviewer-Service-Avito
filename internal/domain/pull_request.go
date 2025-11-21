package domain

import "time"

// пулл реквест
type PullRequest struct {
	PRID string
	PRName string
	AuthorID string
	Status string			 // OPEN / MERGED
	AssignReviewers []*User // назначенные ревьюеры
	CreatedAt time.Time
	MergedAt *time.Time
}