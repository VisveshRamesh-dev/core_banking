CREATE TABLE IF NOT EXISTS accounts (
    id                    BIGSERIAL    PRIMARY KEY,
    customer_id           BIGINT       NOT NULL,
    type                  SMALLINT     NOT NULL,
    status                SMALLINT     NOT NULL DEFAULT 1,
    currency              VARCHAR(8)   NOT NULL,
    cached_balance_minor  BIGINT       NOT NULL DEFAULT 0,
    overdraft_limit_minor BIGINT       NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_customer_id ON accounts (customer_id);
CREATE INDEX IF NOT EXISTS idx_accounts_status      ON accounts (status);
