-- +goose Up
-- +goose StatementBegin
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";
ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username") ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";
ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username")
-- +goose StatementEnd
