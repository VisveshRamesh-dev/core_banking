CREATE TABLE IF NOT EXISTS transactions (
    id                      BIGSERIAL    PRIMARY KEY,
    kind                    SMALLINT     NOT NULL,
    state                   SMALLINT     NOT NULL DEFAULT 1,
    from_account_id         BIGINT,
    to_account_id           BIGINT,
    amount_minor            BIGINT       NOT NULL,
    currency                VARCHAR(8)   NOT NULL,
    idempotency_key         VARCHAR(255) NOT NULL,
    ledger_transaction_id   BIGINT,
    failure_reason          VARCHAR(500),
    source_reference        VARCHAR(255),
    destination_reference   VARCHAR(255),
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at            TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_idempotency   ON transactions (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_transactions_from_account         ON transactions (from_account_id) WHERE from_account_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_to_account           ON transactions (to_account_id)   WHERE to_account_id   IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_ledger_tx            ON transactions (ledger_transaction_id) WHERE ledger_transaction_id IS NOT NULL;
