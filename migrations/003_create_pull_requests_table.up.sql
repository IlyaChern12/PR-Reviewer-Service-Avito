CREATE TABLE IF NOT EXISTS pull_requests (
    pr_id TEXT PRIMARY KEY,
    pr_name TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(user_id),
    status TEXT NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT now(),
    merged_at TIMESTAMP
);