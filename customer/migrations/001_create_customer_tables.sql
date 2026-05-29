-- =============================================================================
-- Customer Service — Initial Schema
-- Migration: 001_create_customer_tables
-- =============================================================================
-- No foreign key constraints; referential integrity is enforced in code.
-- Enum ordinals match proto definitions in common.v1 / customer.v1.
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- individual_customers
-- KYC fields for natural persons. Created first so customers can reference it.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS individual_customers (
    id            BIGSERIAL    PRIMARY KEY,
    date_of_birth DATE,
    nationality   CHAR(2),
    national_id   VARCHAR(100)
);


-- ─────────────────────────────────────────────────────────────────────────────
-- business_customers
-- KYC fields for business entities plus the authorized representative
-- (proprietor) stored inline — one representative per business account.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS business_customers (
    id                  BIGSERIAL    PRIMARY KEY,
    company_name        VARCHAR(255) NOT NULL,
    registration_number VARCHAR(100),
    tax_id              VARCHAR(100),
    prop_first_name     VARCHAR(100) NOT NULL,
    prop_last_name      VARCHAR(100) NOT NULL,
    prop_email          VARCHAR(255) NOT NULL,
    prop_national_id    VARCHAR(100)
);


-- ─────────────────────────────────────────────────────────────────────────────
-- customers
-- Core identity record. One row per customer regardless of type.
-- customer_type: 1=INDIVIDUAL  2=BUSINESS
-- kyc_status:    1=PENDING  2=VERIFIED  3=ACTIVE  4=SUSPENDED  5=CLOSED
-- Exactly one of individual_id / business_id will be set; the other is NULL.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS customers (
    id            BIGSERIAL    PRIMARY KEY,
    customer_type SMALLINT     NOT NULL DEFAULT 0,
    first_name    VARCHAR(100) NOT NULL,
    last_name     VARCHAR(100) NOT NULL,
    full_name     VARCHAR(255),
    email         VARCHAR(255) NOT NULL,
    kyc_status    SMALLINT     NOT NULL DEFAULT 1,
    individual_id BIGINT,
    business_id   BIGINT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_email      ON customers (email);
CREATE        INDEX IF NOT EXISTS idx_customers_kyc_status ON customers (kyc_status);
CREATE        INDEX IF NOT EXISTS idx_customers_type       ON customers (customer_type);


-- ─────────────────────────────────────────────────────────────────────────────
-- phones
-- Canonical phone table shared across all entity types.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS phones (
    id           BIGSERIAL   PRIMARY KEY,
    phone_type   SMALLINT    NOT NULL DEFAULT 0,
    country_code VARCHAR(8)  NOT NULL,
    number       VARCHAR(30) NOT NULL,
    is_primary   BOOLEAN     NOT NULL DEFAULT FALSE
);


-- ─────────────────────────────────────────────────────────────────────────────
-- addresses
-- Canonical address table shared across all entity types.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS addresses (
    id           BIGSERIAL    PRIMARY KEY,
    address_type SMALLINT     NOT NULL DEFAULT 0,
    line1        VARCHAR(255) NOT NULL,
    line2        VARCHAR(255),
    city         VARCHAR(100) NOT NULL,
    state        VARCHAR(100) NOT NULL,
    postal_code  VARCHAR(20)  NOT NULL,
    country      CHAR(2)      NOT NULL,
    is_primary   BOOLEAN      NOT NULL DEFAULT FALSE
);


-- ─────────────────────────────────────────────────────────────────────────────
-- rel_contact
-- Polymorphic junction: links a phone or address record to an owning entity.
--
-- contact_type: 1=PHONE  2=ADDRESS
-- link_type:    1=INDIVIDUAL  (link_id = individual_customers.id)
--               2=BUSINESS    (link_id = business_customers.id, company contacts)
--               3=PROPRIETOR  (link_id = business_customers.id, proprietor phones)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS rel_contact (
    id           BIGSERIAL PRIMARY KEY,
    contact_id   BIGINT    NOT NULL,
    contact_type SMALLINT  NOT NULL,
    link_id      BIGINT    NOT NULL,
    link_type    SMALLINT  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rel_contact_link    ON rel_contact (link_id, link_type);
CREATE INDEX IF NOT EXISTS idx_rel_contact_contact ON rel_contact (contact_id, contact_type);
