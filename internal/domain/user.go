package domain

// участник команды
type User struct {
	UserID string
	Username string
	TeamName string
	IsActive bool
}

// dto
type PullRequestShort struct {
    PRID     string `json:"pull_request_id"`
    PRName   string `json:"pull_request_name"`
    AuthorID string `json:"author_id"`
    Status   string `json:"status"`
}