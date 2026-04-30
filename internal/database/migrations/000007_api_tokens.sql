-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS api_tokens (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    token_hash TEXT NOT NULL,
    token_prefix TEXT NOT NULL,
    last_used_at DATETIME,
    created_at DATETIME,
    revoked_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_token_hash ON api_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_api_tokens_token_prefix ON api_tokens(token_prefix);
CREATE INDEX IF NOT EXISTS idx_api_tokens_revoked_at ON api_tokens(revoked_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_api_tokens_revoked_at;
DROP INDEX IF EXISTS idx_api_tokens_token_prefix;
DROP INDEX IF EXISTS idx_api_tokens_token_hash;
DROP TABLE IF EXISTS api_tokens;
-- +goose StatementEnd
