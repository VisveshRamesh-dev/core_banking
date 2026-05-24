# CoreBank — Architecture & API Reference

A proto-first, multi-module core-banking platform built with Go and Kratos.
This document is the authoritative reference for the system: its architecture,
design principles, and a complete per-RPC specification of every module.

---

## 1. System overview

CoreBank models the core of a retail bank as four cooperating services, each
owning its own data and exposing both gRPC (internal, service-to-service) and
REST/JSON (external, via the gateway) from a single set of `.proto` contracts.

```
                          Gateway (HTTP edge)
                    auth · rate-limit · request-id
                                │ gRPC
        ┌───────────────────────┼────────────────────────┐
        ▼                       ▼                         ▼
   ┌─────────┐            ┌──────────┐             ┌──────────────┐
   │Customer │            │ Account  │             │ Transaction  │
   │ service │            │ service  │             │   service    │
   │         │            │          │             │ (orchestrator)│
   │ KYC     │◀───────────│ refs     │             │ saga over    │
   │ lifecycle│  validate  │ customer │             │ Account+Ledger│
   └────┬────┘            └────┬─────┘             └──────┬───────┘
        │                      │                          │
        ▼                      ▼                          ▼
   customer_db            account_db               ┌──────────────┐
                       (cached balance)            │   Ledger     │
                                                   │   service    │
                                                   │ double-entry │
                                                   │ append-only  │
                                                   │ idempotent   │
                                                   │ SOURCE OF    │
                                                   │ TRUTH        │
                                                   └──────┬───────┘
                                                          ▼
                                                      ledger_db
```

| Service | Responsibility | Owns |
| --- | --- | --- |
| **Customer** | Identity & KYC lifecycle | `customer_db` |
| **Account** | Account lifecycle, types, cached balance | `account_db` |
| **Ledger** | Double-entry money movement (source of truth) | `ledger_db` |
| **Transaction** | Orchestrates business operations across services | orchestration records |

---

## 2. Design principles

These are the decisions that make the system coherent. Each is a deliberate
choice with a defensible rationale.

### 2.1 One-directional dependencies (acyclic)

```
Transaction ──▶ Account ──▶ Customer
            └──▶ Ledger
```

Dependencies only ever point one way. Customer knows nothing about Accounts;
Account knows nothing about Transactions. This prevents circular coupling and
keeps each service independently deployable and testable. It is why the
`Customer` message has no `accounts` field — the reference lives on the
`Account` side (`customer_id`).

### 2.2 Each service owns its data

No service reads another service's database directly. Cross-service data is
fetched over gRPC (e.g. Account calls `CustomerService.GetCustomer` to validate
a customer before opening an account). This keeps schemas private and
independently evolvable, at the cost of network calls — a trade-off discussed
in §7.

### 2.3 Money is always integer minor units

Every monetary amount is an `int64` in the smallest currency unit (paise,
cents) plus an ISO-4217 currency code, expressed through the shared
`common.v1.Money` type. **Floating-point is never used for money** anywhere in
the system, eliminating rounding errors.

### 2.4 The ledger is the single source of truth

Account balances are stored in two places: a **cached** copy on the account
(for fast reads) and the **authoritative** value derived from the ledger's
entries. They are reconciled; on any disagreement, the ledger wins.

### 2.5 The ledger is append-only

Ledger entries are immutable. There is no update or delete. Corrections are
made by posting a new, reversing transaction, preserving a complete and
auditable history. This is what makes the system trustworthy to auditors.

### 2.6 Double-entry invariant

Every ledger transaction consists of two or more entries whose signed amounts
**sum to zero**. Money is never created or destroyed, only moved. An unbalanced
transaction is rejected before any balance changes.

### 2.7 Idempotency on every money-moving operation

All write operations that move money accept an `idempotency_key`. Re-submitting
the same key returns the original result instead of applying the operation
twice — this makes client retries (after a timeout) safe, which is essential in
payments where the network is unreliable.

### 2.8 Lifecycle as a state machine

Customers and accounts each have a status enum with **validated transitions**.
Status changes are not arbitrary `UPDATE`s — illegal moves (e.g. `CLOSED →
ACTIVE`) are rejected by the business layer.

---

## 3. Common types (`common.v1`)

Shared across all services so a concept means the same thing everywhere.

### Money
| Field | Type | Description |
| --- | --- | --- |
| `amount_minor` | `int64` | Amount in the smallest indivisible unit (e.g. ₹1000.00 → `100000`). |
| `currency` | `string` | ISO-4217 code, e.g. `INR`, `USD`. |

### PageRequest / PageResponse
| Field | Type | Description |
| --- | --- | --- |
| `page_size` | `int32` | Max items to return (server caps it). `0` → server default. |
| `page_token` | `string` | Opaque cursor from a previous response. Empty → first page. |
| `next_page_token` | `string` | Cursor for the next page. Empty → no more pages. |
| `total_size` | `int64` | Total matching records (optional). |

### ErrorReason
Machine-readable error codes returned in error metadata so clients branch on a
stable enum rather than parsing messages.

| Reason | When |
| --- | --- |
| `VALIDATION_FAILED` | Request failed field validation. |
| `NOT_FOUND` | Entity does not exist. |
| `ALREADY_EXISTS` | Unique-constraint conflict. |
| `PERMISSION_DENIED` | Caller not authorised. |
| `INTERNAL` | Unexpected server error. |
| `CUSTOMER_NOT_ACTIVE` | Account operation attempted on a non-active customer. |
| `ACCOUNT_FROZEN` | Money operation on a frozen account. |
| `ACCOUNT_CLOSED` | Money operation on a closed account. |
| `INSUFFICIENT_FUNDS` | Debit exceeds balance + overdraft. |
| `CURRENCY_MISMATCH` | Entries span multiple currencies. |
| `UNBALANCED_ENTRIES` | Ledger entries do not sum to zero. |
| `ILLEGAL_TRANSITION` | Invalid status state-machine move. |
| `IDEMPOTENCY_CONFLICT` | Same idempotency key submitted with a different payload. |

---

## 4. Customer service

Manages the customer identity lifecycle. A customer is **onboarded** through a
KYC state machine and is **never hard-deleted** (regulatory record retention).

### KYC state machine
```
PENDING ──▶ VERIFIED ──▶ ACTIVE ──▶ SUSPENDED ──▶ ACTIVE
                            │
                            └──▶ CLOSED  (terminal)
```

### 4.1 Onboard
Creates a new customer in `PENDING` status.

- **HTTP:** `POST /v1/customers`
- **Request:** `type`, `full_name`, `email`, `phone`
- **Response:** `Customer` (status = `PENDING`)
- **Rules:** email/phone format validated; duplicate email → `ALREADY_EXISTS`.
- **Errors:** `VALIDATION_FAILED`, `ALREADY_EXISTS`.

### 4.2 GetCustomer
Fetches a single customer by ID.

- **HTTP:** `GET /v1/customers/{id}`
- **Request:** `id`
- **Response:** `Customer`
- **Errors:** `NOT_FOUND`.

### 4.3 UpdateKYCStatus
Transitions a customer through the KYC lifecycle. **Not** a generic update —
the target transition is validated against the state machine.

- **HTTP:** `PATCH /v1/customers/{id}/kyc-status`
- **Request:** `id`, `new_status`, `reason` (audit note)
- **Response:** `Customer` (updated status)
- **Rules:** transition must be legal; e.g. `CLOSED → ACTIVE` is rejected.
- **Errors:** `NOT_FOUND`, `ILLEGAL_TRANSITION`, `VALIDATION_FAILED`.

### 4.4 CloseCustomer
Moves a customer to the terminal `CLOSED` status (soft close — no deletion).

- **HTTP:** `POST /v1/customers/{id}:close`
- **Request:** `id`, `reason`
- **Response:** `Customer` (status = `CLOSED`)
- **Rules:** typically requires no active accounts (enforced by orchestration).
- **Errors:** `NOT_FOUND`, `ILLEGAL_TRANSITION`.

### 4.5 ListCustomers
Paginated list, optionally filtered by status.

- **HTTP:** `GET /v1/customers`
- **Request:** `status_filter` (optional), `page`
- **Response:** `customers[]`, `page`
- **Errors:** `VALIDATION_FAILED`.

---

## 5. Account service

Manages account lifecycle. Depends on Customer (validates the owner is active)
but is never depended upon by Customer.

### Account state machine
```
PENDING ──▶ ACTIVE ──▶ FROZEN ──▶ ACTIVE
               │
               └──▶ CLOSED  (terminal; requires zero balance)
```

### Account types
| Type | Behaviour |
| --- | --- |
| `SAVINGS` | Balance cannot go below zero. |
| `CURRENT` | May permit an overdraft up to `overdraft_limit_minor`. |
| `WALLET` | Prepaid; balance only. |
| `LOAN` | Balance represents outstanding principal. |

### 5.1 OpenAccount
Creates an account for an existing, active customer.

- **HTTP:** `POST /v1/accounts`
- **Request:** `customer_id`, `type`, `currency`, `overdraft_limit_minor` (optional)
- **Response:** `Account` (status = `PENDING` or `ACTIVE`)
- **Rules:** calls `CustomerService.GetCustomer`; the customer must exist and be
  `ACTIVE`. Currency is fixed at open time. Overdraft only honoured for `CURRENT`.
- **Errors:** `NOT_FOUND` (customer), `CUSTOMER_NOT_ACTIVE`, `VALIDATION_FAILED`.

### 5.2 GetAccount
Fetches an account by ID (includes cached balance).

- **HTTP:** `GET /v1/accounts/{id}`
- **Request:** `id`
- **Response:** `Account`
- **Errors:** `NOT_FOUND`.

### 5.3 ListAccountsByCustomer
All accounts belonging to one customer.

- **HTTP:** `GET /v1/customers/{customer_id}/accounts`
- **Request:** `customer_id`, `page`
- **Response:** `accounts[]`, `page`
- **Errors:** `NOT_FOUND` (customer).

### 5.4 UpdateAccountStatus
Transitions an account (activate, freeze, unfreeze). Validated against the
state machine.

- **HTTP:** `PATCH /v1/accounts/{id}/status`
- **Request:** `id`, `new_status`, `reason`
- **Response:** `Account`
- **Errors:** `NOT_FOUND`, `ILLEGAL_TRANSITION`.

### 5.5 CloseAccount
Moves an account to `CLOSED`. Rejected unless the balance is zero.

- **HTTP:** `POST /v1/accounts/{id}:close`
- **Request:** `id`, `reason`
- **Response:** `Account` (status = `CLOSED`)
- **Rules:** authoritative balance (from Ledger) must be zero.
- **Errors:** `NOT_FOUND`, `ILLEGAL_TRANSITION`, `VALIDATION_FAILED` (non-zero balance).

---

## 6. Ledger service

The source of truth for money. Small by design: post, reverse, read a
transaction, read a balance. **No update or delete** — it is append-only.

### Core concepts
- **Entry:** one signed leg against one account (`DEBIT` decreases, `CREDIT`
  increases). `amount_minor` is always positive; direction carries the sign.
- **LedgerTransaction:** an atomic set of entries that sum to zero.
- **Balance:** derived from the sum of an account's posted entries.

### 6.1 PostTransaction
Atomically applies a balanced set of entries.

- **HTTP:** `POST /v1/ledger/transactions`
- **Request:** `entries[]`, `idempotency_key`, `description`
- **Response:** `LedgerTransaction` (status = `POSTED` or `REJECTED`)
- **Validation order (all-or-nothing):**
  1. entries sum to zero → else `UNBALANCED_ENTRIES`
  2. single shared currency → else `CURRENCY_MISMATCH`
  3. no account `FROZEN`/`CLOSED` → else `ACCOUNT_FROZEN` / `ACCOUNT_CLOSED`
  4. debited accounts have sufficient funds (incl. overdraft) → else `INSUFFICIENT_FUNDS`
- **Idempotency:** re-posting the same key returns the original transaction;
  same key + different payload → `IDEMPOTENCY_CONFLICT`.
- **Atomicity:** applied within a single DB transaction; on any failure nothing
  is written and status is `REJECTED`.

### 6.2 ReverseTransaction
Posts a **new** transaction with the opposite entries of an existing one. The
original is never modified.

- **HTTP:** `POST /v1/ledger/transactions/{transaction_id}:reverse`
- **Request:** `transaction_id`, `reason`, `idempotency_key`
- **Response:** `LedgerTransaction` (the new reversing transaction, with
  `reverses_transaction_id` set)
- **Errors:** `NOT_FOUND`, `IDEMPOTENCY_CONFLICT`.

### 6.3 GetTransaction
Reads a ledger transaction and its entries.

- **HTTP:** `GET /v1/ledger/transactions/{transaction_id}`
- **Request:** `transaction_id`
- **Response:** `LedgerTransaction`
- **Errors:** `NOT_FOUND`.

### 6.4 GetBalance
Returns the authoritative balance for an account.

- **HTTP:** `GET /v1/ledger/accounts/{account_id}/balance`
- **Request:** `account_id`
- **Response:** `Balance` (`account_id`, `Money`, `as_of` timestamp)
- **Errors:** `NOT_FOUND`.

---

## 7. Transaction service (orchestrator)

Owns no balances. Translates a business intent ("transfer ₹100") into ledger
entries and coordinates Account + Ledger, guaranteeing all-or-nothing semantics
across service boundaries. This is the saga layer.

### Orchestration lifecycle
```
PENDING ──▶ COMPLETED   (ledger posted)
        └─▶ FAILED      (rejected or compensated; no net money moved)
```

### 7.1 Transfer
Account-to-account money movement.

- **HTTP:** `POST /v1/transactions/transfer`
- **Request:** `from_account_id`, `to_account_id`, `amount_minor`, `currency`,
  `idempotency_key`, `description`
- **Response:** `Transaction` (with `ledger_transaction_id` once posted)
- **Orchestration:**
  1. validate both accounts exist and are `ACTIVE` (Account service)
  2. post a balanced `DEBIT from / CREDIT to` to the Ledger
  3. ledger success → `COMPLETED`; failure → `FAILED` (nothing moved)
- **Errors:** `NOT_FOUND`, `ACCOUNT_FROZEN`, `INSUFFICIENT_FUNDS`,
  `CURRENCY_MISMATCH`, `IDEMPOTENCY_CONFLICT`.

### 7.2 Deposit
Credits an account from an external source.

- **HTTP:** `POST /v1/transactions/deposit`
- **Request:** `to_account_id`, `amount_minor`, `currency`, `idempotency_key`,
  `source_reference`
- **Response:** `Transaction`
- **Orchestration:** ledger entry credits the account and debits a system
  "external/clearing" account so the transaction still balances.
- **Errors:** `NOT_FOUND`, `ACCOUNT_FROZEN`, `ACCOUNT_CLOSED`.

### 7.3 Withdraw
Debits an account to an external destination.

- **HTTP:** `POST /v1/transactions/withdraw`
- **Request:** `from_account_id`, `amount_minor`, `currency`, `idempotency_key`,
  `destination_reference`
- **Response:** `Transaction`
- **Errors:** `NOT_FOUND`, `ACCOUNT_FROZEN`, `INSUFFICIENT_FUNDS`.

### 7.4 GetTransaction
Returns the orchestration record and its state.

- **HTTP:** `GET /v1/transactions/{id}`
- **Request:** `id`
- **Response:** `Transaction`
- **Errors:** `NOT_FOUND`.

### 7.5 ListAccountTransactions
Transaction history for an account.

- **HTTP:** `GET /v1/accounts/{account_id}/transactions`
- **Request:** `account_id`, `page`
- **Response:** `transactions[]`, `page`
- **Errors:** `NOT_FOUND` (account).

---

## 8. Cross-cutting concerns

### 8.1 Shared middleware (`pkg/middleware`)
Implemented once, imported by every service's `main.go`, applied as a chain:
```
recovery → request-id → logging → tracing → auth → rate-limit → [handler]
```
The request-id generated at the gateway propagates through gRPC metadata to
every downstream service, so one trace follows a request across all modules.

### 8.2 Idempotency (`pkg/idempotency`)
Money-moving handlers persist `(idempotency_key → result)`. On a repeat key
with an identical payload, the stored result is returned; with a differing
payload, `IDEMPOTENCY_CONFLICT` is raised.

### 8.3 Cross-service failure handling
Because Account and Ledger are separate services, a `Transfer` can fail
mid-orchestration. Options the system is designed to support: fail-closed
(reject if Account validation can't be confirmed), short-TTL caching of account
status, and compensating actions (the Transaction state machine records
`FAILED` and ensures no ledger entry persists). The ledger's atomic
`PostTransaction` means the actual money step is itself all-or-nothing.

---

## 9. End-to-end example: a ₹250 transfer

```
1. POST /v1/transactions/transfer
     { from: acc_alice, to: acc_bob, amount_minor: 25000, currency: INR,
       idempotency_key: "txn-abc-123" }

2. Transaction service:
     a. AccountService.GetAccount(acc_alice)  → ACTIVE, INR ✓
     b. AccountService.GetAccount(acc_bob)    → ACTIVE, INR ✓
     c. LedgerService.PostTransaction(
          entries = [ {acc_alice, DEBIT,  25000, INR},
                      {acc_bob,   CREDIT, 25000, INR} ],
          idempotency_key = "txn-abc-123")
            → validates sum == 0 ✓, currency ✓, not frozen ✓, funds ✓
            → POSTED
     d. state = COMPLETED, ledger_transaction_id set

3. Response: Transaction { state: COMPLETED, ledger_transaction_id: ... }

If step 2c fails (e.g. INSUFFICIENT_FUNDS), state = FAILED and no money moves.
A client retry with the same idempotency_key returns the original result.
```

---

## 10. Proto layout

```
api/
├── common/v1/      common.proto (Money, pagination)  ·  error.proto (ErrorReason)
├── customer/v1/    customer.proto (types)  ·  customer_service.proto (rpcs)
├── account/v1/     account.proto           ·  account_service.proto
├── ledger/v1/      ledger.proto            ·  ledger_service.proto
└── transaction/v1/ transaction.proto       ·  transaction_service.proto
```

Each module separates **types** (messages/enums) from **service** (RPCs) so a
service can import another's types without importing its endpoints. Everything
is versioned under `v1` from day one; breaking changes introduce `v2` alongside,
never mutate `v1`.
