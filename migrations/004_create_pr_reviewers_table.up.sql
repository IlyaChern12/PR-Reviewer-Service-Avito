CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pr_id TEXT REFERENCES pull_requests(pr_id),
    user_id TEXT REFERENCES users(user_id),
    PRIMARY KEY (pr_id, user_id)
);