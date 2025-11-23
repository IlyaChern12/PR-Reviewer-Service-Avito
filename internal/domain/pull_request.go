package domain

import "time"

// пулл реквест
type PullRequest struct {
	PRID            string     `json:"pull_request_id" db:"pull_request_id"`
	PRName          string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID        string     `json:"author_id" db:"author_id"`
	Status          string     `json:"status" db:"status"`           // OPEN / MERGED
	AssignReviewers []*User    `json:"assigned_reviewers,omitempty"` // назначенные ревьюеры
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	MergedAt        *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

// сокращенный PR (dto)
type PullRequestShort struct {
	PRID     string `json:"pull_request_id"`
	PRName   string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}
