package queries

// UserRepo
const (
	InsertOrUpdateUser = `
		INSERT INTO users(user_id, username, team_name, is_active)
		VALUES($1, $2, $3, $4)
		ON CONFLICT(user_id) DO UPDATE
		SET username = EXCLUDED.username,
		is_active = EXCLUDED.is_active,
		team_name = EXCLUDED.team_name`

	SelectUserByID = `
		SELECT user_id, username, team_name, is_active FROM users
		WHERE user_id=$1`

	UpdateUserIsActive = `
		UPDATE users SET is_active=$1
		WHERE user_id=$2`

	SelectUsersByTeam        = `
		SELECT user_id, username, team_name, is_active FROM users
		WHERE team_name=$1`

	SelectActiveUsersByTeam  = `
		SELECT user_id, username, team_name, is_active FROM users
		WHERE team_name=$1 AND is_active=true`

	SelectReviewPRsByUser = `
		SELECT pr.pull_request_id AS pr_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_request_reviewers prr
		JOIN pull_requests pr ON prr.pull_request_id = pr.pull_request_id
		WHERE prr.user_id = $1;`
)

// TeamRepo
const (
	InsertTeam      = `
		INSERT INTO teams(team_name) VALUES($1)`

	SelectTeamExist = `
		SELECT EXISTS(SELECT 1 FROM teams
		WHERE team_name=$1)`

	SelectTeamUsers = `
		SELECT user_id, username, is_active FROM users
		WHERE team_name=$1`
)

// PullRequestRepo
const (
	SelectAllRPs = `
		SELECT pull_request_id FROM pull_requests`

	SelectPRExist           = `
		SELECT EXISTS(SELECT 1 FROM pull_requests
		WHERE pull_request_id=$1)`

	InsertPR                = `
		INSERT INTO pull_requests(pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES($1,$2,$3,'OPEN',NOW())`

	UpdatePRStatusMerged    = `
		UPDATE pull_requests SET status='MERGED', merged_at=$1
		WHERE pull_request_id=$2`

	SelectPRByID            = `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id=$1`

	SelectPRReviewersFull = `
		SELECT u.user_id, u.username, u.team_name, u.is_active
		FROM pull_request_reviewers prr
		JOIN users u ON prr.user_id = u.user_id
		WHERE prr.pull_request_id = $1`

	SelectPRReviewers       = `
		SELECT COUNT(1) FROM pull_request_reviewers
		WHERE pull_request_id=$1 AND user_id=$2`

	InsertPRReviewer = `
		INSERT INTO pull_request_reviewers(pull_request_id, user_id)
		VALUES($1, $2)`

	UpdatePRReviewer        = `
		UPDATE pull_request_reviewers SET user_id=$1
		WHERE pull_request_id=$2 AND user_id=$3`

	SelectPRsByReviewer = `
		SELECT pr.pull_request_id AS pr_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id=$1;`
)