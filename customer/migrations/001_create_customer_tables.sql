-- =============================================================================
-- Customer Service — Initial Schema
-- Migration: 001_create_customer_tables
-- =============================================================================
-- No foreign key constraints are defined here; referential integrity is
-- enforced in application code. All relational columns are plain BIGINT.
-- Enum values match the proto enum ordinals in common.v1 / customer.v1.
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- customers
-- Core identity record. One row per customer regardless of type.
-- customer_type: 1=INDIVIDUAL  2=BUSINESS
-- kyc_status:    1=PENDING  2=VERIFIED  3=ACTIVE  4=SUSPENDED  5=CLOSED
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS customers (
    id            BIGSERIAL    PRIMARY KEY,
    customer_type SMALLINT     NOT NULL DEFAULT 0,
    first_name    VARCHAR(100) NOT NULL,
    last_name     VARCHAR(100) NOT NULL,
    full_name     VARCHAR(255),
    email         VARCHAR(255) NOT NULL,
    kyc_status    SMALLINT     NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_email     ON customers (email);
CREATE        INDEX IF NOT EXISTS idx_customers_kyc_status ON customers (kyc_status);
CREATE        INDEX IF NOT EXISTS idx_customers_type       ON customers (customer_type);


-- ─────────────────────────────────────────────────────────────────────────────
-- customer_phones
-- Repeated common.v1.Phone on the Customer message.
-- phone_type: 1=MOBILE  2=LANDLINE  3=WORK  4=FAX
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS customer_phones (
    id           BIGSERIAL    PRIMARY KEY,
    customer_id  BIGINT       NOT NULL,
    phone_type   SMALLINT     NOT NULL DEFAULT 0,
    country_code VARCHAR(8)   NOT NULL,
    number       VARCHAR(30)  NOT NULL,
    is_primary   BOOLEAN      NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_customer_phones_customer_id ON customer_phones (customer_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- customer_addresses
-- Repeated common.v1.Address on the Customer message.
-- address_type: 1=HOME  2=WORK  3=REGISTERED  4=CORRESPONDENCE
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS customer_addresses (
    id           BIGSERIAL    PRIMARY KEY,
    customer_id  BIGINT       NOT NULL,
    address_type SMALLINT     NOT NULL DEFAULT 0,
    line1        VARCHAR(255) NOT NULL,
    line2        VARCHAR(255),
    city         VARCHAR(100) NOT NULL,
    state        VARCHAR(100) NOT NULL,
    postal_code  VARCHAR(20)  NOT NULL,
    country      CHAR(2)      NOT NULL,
    is_primary   BOOLEAN      NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_customer_addresses_customer_id ON customer_addresses (customer_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- customer_individual_details
-- KYC fields for INDIVIDUAL customers (customer_type = 1).
-- One row per customer; customer_id references customers.id in code.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS customer_individual_details (
    id            BIGSERIAL   PRIMARY KEY,
    customer_id   BIGINT      NOT NULL,
    date_of_birth DATE,
    nationality   CHAR(2),
    national_id   VARCHAR(100)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_individual_details_customer_id ON customer_individual_details (customer_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- customer_business_details
-- KYC fields for BUSINESS customers (customer_type = 2).
-- One row per customer; customer_id references customers.id in code.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS customer_business_details (
    id                  BIGSERIAL    PRIMARY KEY,
    customer_id         BIGINT       NOT NULL,
    company_name        VARCHAR(255) NOT NULL,
    registration_number VARCHAR(100),
    tax_id              VARCHAR(100)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_details_customer_id ON customer_business_details (customer_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- business_phones
-- Repeated common.v1.Phone on the BusinessDetails message (company_phones).
-- business_id references customer_business_details.id in code.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS business_phones (
    id           BIGSERIAL   PRIMARY KEY,
    business_id  BIGINT      NOT NULL,
    phone_type   SMALLINT    NOT NULL DEFAULT 0,
    country_code VARCHAR(8)  NOT NULL,
    number       VARCHAR(30) NOT NULL,
    is_primary   BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_business_phones_business_id ON business_phones (business_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- business_addresses
-- Repeated common.v1.Address on the BusinessDetails message (registered_addresses).
-- business_id references customer_business_details.id in code.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS business_addresses (
    id           BIGSERIAL    PRIMARY KEY,
    business_id  BIGINT       NOT NULL,
    address_type SMALLINT     NOT NULL DEFAULT 0,
    line1        VARCHAR(255) NOT NULL,
    line2        VARCHAR(255),
    city         VARCHAR(100) NOT NULL,
    state        VARCHAR(100) NOT NULL,
    postal_code  VARCHAR(20)  NOT NULL,
    country      CHAR(2)      NOT NULL,
    is_primary   BOOLEAN      NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_business_addresses_business_id ON business_addresses (business_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- business_proprietors
-- ProprietorInfo nested inside BusinessDetails.
-- business_id references customer_business_details.id in code.
-- One row per business (one authorised representative per account).
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS business_proprietors (
    id          BIGSERIAL    PRIMARY KEY,
    business_id BIGINT       NOT NULL,
    first_name  VARCHAR(100) NOT NULL,
    last_name   VARCHAR(100) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    national_id VARCHAR(100)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_proprietors_business_id ON business_proprietors (business_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- proprietor_phones
-- Repeated common.v1.Phone on the ProprietorInfo message.
-- proprietor_id references business_proprietors.id in code.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS proprietor_phones (
    id            BIGSERIAL   PRIMARY KEY,
    proprietor_id BIGINT      NOT NULL,
    phone_type    SMALLINT    NOT NULL DEFAULT 0,
    country_code  VARCHAR(8)  NOT NULL,
    number        VARCHAR(30) NOT NULL,
    is_primary    BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_proprietor_phones_proprietor_id ON proprietor_phones (proprietor_id);
