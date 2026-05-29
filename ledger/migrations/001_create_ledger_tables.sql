-- =============================================================================
-- Ledger Service — Initial Schema
-- Migration: 001_create_ledger_tables
-- =============================================================================
-- The ledger is append-only: no UPDATE or DELETE on entries ever.
-- Mistakes are corrected by posting a new reversing transaction.
-- status: 1=POSTED  2=REJECTED
-- direction: 1=DEBIT  2=CREDIT
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- ledger_transactions
-- One row per atomic money movement. The idempotency_key enforces exactly-once
-- semantics — re-posting the same key returns the original row.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ledger_transactions (
    id                      BIGSERIAL    PRIMARY KEY,
    status                  SMALLINT     NOT NULL DEFAULT 1,
    description             VARCHAR(500),
    idempotency_key         VARCHAR(255) NOT NULL,
    reverses_transaction_id BIGINT,
    posted_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ledger_transactions_idempotency
    ON ledger_transactions (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_ledger_transactions_reverses
    ON ledger_transactions (reverses_transaction_id)
    WHERE reverses_transaction_id IS NOT NULL;


-- ─────────────────────────────────────────────────────────────────────────────
-- ledger_entries
-- Each row is one leg of a transaction (DEBIT or CREDIT against one account).
-- The signed sum of all entries in a transaction must equal zero.
-- amount_minor is always positive; direction carries the sign.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ledger_entries (
    id             BIGSERIAL   PRIMARY KEY,
    transaction_id BIGINT      NOT NULL,
    account_id     BIGINT      NOT NULL,
    direction      SMALLINT    NOT NULL,
    amount_minor   BIGINT      NOT NULL,
    currency       VARCHAR(8)  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ledger_entries_transaction_id
    ON ledger_entries (transaction_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_account_id
    ON ledger_entries (account_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_account_currency
    ON ledger_entries (account_id, currency);
