-- +goose Up
-- +goose StatementBegin
-- NOTE: Default value 'data/temp' must match config.DefaultTempDir constant in internal/config/config.go
ALTER TABLE jobs ADD COLUMN temp_dir TEXT NOT NULL DEFAULT 'data/temp';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite doesn't support DROP COLUMN directly, so we recreate the table
CREATE TABLE jobs_backup (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    total_files INTEGER NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    progress REAL NOT NULL DEFAULT 0,
    destination TEXT NOT NULL DEFAULT '',
    files TEXT NOT NULL,
    results TEXT NOT NULL DEFAULT '{}',
    excluded TEXT NOT NULL DEFAULT '{}',
    file_match_info TEXT NOT NULL DEFAULT '{}',
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    organized_at DATETIME
);

INSERT INTO jobs_backup SELECT id, status, total_files, completed, failed, progress, destination, files, results, excluded, file_match_info, started_at, completed_at, organized_at FROM jobs;

DROP TABLE jobs;

ALTER TABLE jobs_backup RENAME TO jobs;
-- +goose StatementEnd
