CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id TEXT NOT NULL,
    status TEXT DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT now(),
    merged_at TIMESTAMP
);