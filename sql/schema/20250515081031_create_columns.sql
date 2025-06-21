-- +goose Up
-- +goose StatementBegin

CREATE TABLE accounts(
    id uuid PRIMARY KEY default gen_random_uuid(),
    "owner" varchar(256) NOT NULL,
    balance numeric(9,2) NOT NULL,
    currency varchar(100) NOT NULL,
    created_at timestamptz NOT NULL default now(),
    updated_at timestamptz NOT NULL default now()
);
CREATE INDEX idx_accounts_owner ON accounts("owner");


CREATE TABLE transfers(
    id uuid PRIMARY KEY default gen_random_uuid(),
    from_account_id uuid NOT NULL,
    to_account_id uuid NOT NULL,
    amount numeric(9,2) NOT NULL,
    created_at timestamptz NOT NULL default now(),
    updated_at timestamptz NOT NULL default now(),

    CONSTRAINT amount_positive CHECK (amount > 0),
    CONSTRAINT different_accounts CHECK (from_account_id != to_account_id),
    CONSTRAINT fk_from_account FOREIGN KEY (from_account_id) REFERENCES accounts(id) ON DELETE cascade,
    CONSTRAINT fk_to_account FOREIGN KEY (to_account_id) REFERENCES accounts(id) ON DELETE cascade

);

CREATE INDEX idx_transfers_from_account_id ON transfers(from_account_id);
CREATE INDEX idx_transfers_to_account_id ON transfers(to_account_id);
CREATE INDEX idx_transfers_from_to_account_id ON transfers(from_account_id,to_account_id);

CREATE TABLE entries(
    id uuid PRIMARY KEY default gen_random_uuid(),
    account_id uuid NOT NULL,
    amount numeric(9,2) NOT NULL,
    created_at timestamptz default now(),
    updated_at timestamptz default now(),

    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES accounts(id) on DELETE cascade,
    CONSTRAINT amount_not_zero CHECK (amount != 0)
);

CREATE INDEX idx_entries_account_id ON entries(account_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS accounts;


-- +goose StatementEnd
