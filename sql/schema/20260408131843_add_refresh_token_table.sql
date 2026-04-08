-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD refresh_token TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_users_refresh_token ON users(refresh_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_refresh_token;
ALTER TABLE users DROP COLUMN IF EXISTS refresh_token;
-- +goose StatementEnd
