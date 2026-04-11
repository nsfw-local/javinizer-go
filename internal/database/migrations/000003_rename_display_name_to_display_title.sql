-- +goose Up
-- +goose StatementBegin
ALTER TABLE movies RENAME COLUMN display_name TO display_title;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE movies RENAME COLUMN display_title TO display_name;
-- +goose StatementEnd
