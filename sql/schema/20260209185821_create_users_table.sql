-- +goose Up
-- +goose StatementBegin
CREATE TABLE "users" (
    "username" varchar PRIMARY KEY,
    "hashed_password" varchar NOT NULL,
    "full_name" varchar NOT NULL,
    "email" varchar UNIQUE NOT NULL,
    "password_changed_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP

);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");
ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE ("owner","currency");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";
DROP TABLE "users";
-- +goose StatementEnd
