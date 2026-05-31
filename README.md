# CoreBank

A proto-first, multi-service core-banking platform in Go — built to model the real
mechanics of how money moves through a bank: customer onboarding, account
products (savings, current, loans), and a double-entry ledger that is the
single source of truth for every paisa.

This is a **portfolio project**: I'm using it as the canvas for designing, in
the open, the kind of system I've worked on in production fintech. The
architecture, design decisions, and trade-offs are documented as carefully as
the code itself.

> 📖 **Full design reference:** [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
> — every service, every RPC, every rule, with the reasoning behind each one.

---

## What CoreBank does

```
A customer is onboarded → passes KYC → opens a "Salary Savings" account
(a configured product from the catalog) → receives a deposit → transfers
money to another customer's account → every entry recorded in an
append-only double-entry ledger → balances always reconcile to the rupee.
```

Five cooperating services, each owning its own data, each exposing both gRPC
(internal) and REST (external) from a single set of protobuf contracts:

| Service       | Owns                                              |
| ------------- | ------------------------------------------------- |
| `product`     | Product catalog (account & loan templates)        |
| `customer`    | Customer identity & KYC lifecycle                 |
| `account`     | Account lifecycle, types, cached balance          |
| `ledger`      | Double-entry, append-only — **source of truth**   |
| `transaction` | Orchestrator: saga across account + ledger        |

---

## Architecture at a glance

```
                          Gateway (HTTP edge)
                    auth · rate-limit · request-id
                                │ gRPC
        ┌───────────────────────┼────────────────────────┐
        ▼                       ▼                         ▼
   ┌─────────┐            ┌──────────┐             ┌──────────────┐
   │Customer │            │ Account  │             │ Transaction  │
   │         │            │ (refs    │             │  (saga over  │
   │ KYC     │◀───────────│ customer │             │   account +  │
   │ states  │  validate  │ + product)│            │    ledger)   │
   └─────────┘            └──────────┘             └──────┬───────┘
                                ▲                          │
                                │                          ▼
                          ┌─────┴────┐              ┌──────────────┐
                          │ Product  │              │    Ledger    │
                          │ catalog  │              │ double-entry │
                          │  (oneof  │              │ append-only  │
                          │  config) │              │ idempotent   │
                          └──────────┘              └──────────────┘
```

Dependencies point one way only. Each service owns its own database. No
foreign keys cross service boundaries (and even within a service, we keep
schemas FK-free for shard-friendliness — integrity lives in the application
layer).

---

## Design decisions worth defending

These are the choices that turn "another microservices demo" into something
that resembles a real bank. Each is documented in detail in
[`ARCHITECTURE.md`](docs/ARCHITECTURE.md).

- **Proto-first, dual transport.** Every RPC carries a
  `google.api.http` annotation, so `buf generate` produces *both* a gRPC
  server (internal) and a REST gateway (external) from one contract. Define
  once, serve twice.

- **Money is integer minor units, always.** `int64` paise/cents +
  ISO-4217 currency. Interest rates in basis points. Floats never touch
  monetary values anywhere in the codebase.

- **Double-entry ledger.** Every transaction has ≥ 2 entries whose
  signed amounts sum to zero. The invariant is checked before any balance
  changes — unbalanced transactions are rejected outright. Money can't be
  created or destroyed by definition.

- **Append-only.** Ledger entries are immutable. Mistakes are
  corrected by posting a reversing transaction, never by editing history.
  Auditors can replay every movement that ever happened.

- **Cached vs. authoritative balance.** Accounts cache a balance for
  fast reads; the ledger holds the truth. The two are reconciled.

- **Products are data, not code.** Account "types" (salary, kids,
  zero-balance, home loan, etc.) live in a configurable catalog with class-
  specific configuration via protobuf `oneof`. Launching a new product is a
  catalog entry, not a deployment.

- **Idempotency on every money-moving RPC.** A unique
  `idempotency_key` enforced at the database level makes "the client
  retried after a timeout" impossible to turn into a double-charge.

- **State-machine lifecycle.** Customer KYC and account status changes
  are validated transitions — not arbitrary updates. Illegal moves
  (`CLOSED → ACTIVE`) are rejected before they touch the database.

- **Saga orchestration.** Cross-service transactions are coordinated
  by an orchestrator with explicit `PENDING → COMPLETED / FAILED` states. A
  `FAILED` saga guarantees no net money moved.

---

## Skills demonstrated

A scannable summary for anyone (recruiters, hiring managers) who'd rather
not read the whole thing:

**Languages & frameworks**
- Go 1.22+ · Kratos v2 · gRPC · Protocol Buffers · buf · GORM

**Backend & systems design**
- Microservices with strict service boundaries and one-way dependencies
- Proto-first contract design with dual HTTP+gRPC transport
- Domain-driven design (state machines, immutable history, product catalog)
- Distributed-transaction patterns (saga, compensation, idempotency)
- Append-only / event-sourcing-flavored ledger design
- Multi-database architecture (one DB per service, no cross-service FKs)

**Data & correctness**
- PostgreSQL schema design with DB-level invariants (CHECK constraints,
  uniqueness for idempotency)
- Money handled as integer minor units throughout (no floats)
- Double-entry bookkeeping correctness under concurrency

**Concurrency & reliability**
- Bounded worker pools with backpressure and graceful shutdown (see the
  companion repo, [PayFlow](#related-projects))
- Race-tested concurrent code (`go test -race`)
- Cross-service failure handling with explicit state machines

**Domain knowledge (fintech)**
- Core banking primitives — customers, accounts, ledgers, products, fees
- Payment lifecycle, idempotency, reversals, reconciliation
- PCI DSS-aware data handling, ISO 8583 message processing
  *(from prior production work — see resume)*

---

## Repository layout

```
corebank/
├── README.md                 ← you are here
├── CLAUDE.md                 ← AI-assist context (architecture rules, conventions)
├── docs/
│   └── ARCHITECTURE.md       ← full design reference (read this!)
├── proto/                    ← .proto contracts + generated Go code
│   ├── common/v1/            ← Money, pagination, ErrorReason
│   ├── product/v1/
│   ├── customer/v1/
│   ├── account/v1/
│   ├── ledger/v1/
│   └── transaction/v1/
├── common/                   ← shared libs: middleware, money helpers, idempotency
├── account/                  ← Kratos service (one per module)
│   └── internal/{service,biz,data,server}
├── customer/
├── product/
├── ledger/
├── transaction/
├── db/
│   └── schema.sql            ← Postgres DDL for all services
└── docker-compose.yml        ← local dev stack (Postgres + services)
```

Each module follows the Kratos `service / biz / data / server` layering —
thin transport (proto ↔ domain mapping), business logic with domain rules,
and a repository layer behind an interface.

---

## Running locally

You'll need: Docker, Go 1.22+, `buf`, and `protoc`.

```bash
# 1. Start Postgres (one container, multiple databases on first boot)
docker compose up -d db

# 2. Generate Go code from the protos
buf generate

# 3. Run a service (example: ledger)
go run ./ledger/cmd/...
```

The ledger listens on `:9000` (gRPC) and `:8000` (HTTP). See each module's
README for service-specific endpoints, or `docs/ARCHITECTURE.md` for the
full API reference.

---

## Status & roadmap

This is built in phases. The tracker lives in `docs/ROADMAP.xlsx`.

- ✅ **Phase 0:** Proto contracts for all 5 modules
- ✅ **Phase 0:** Architecture documentation
- ✅ **Phase 0:** Database schema with no-FK / relationship-table approach
- 🚧 **Phase 1:** Proto codegen pipeline (buf)
- 🚧 **Phase 2:** Ledger service end-to-end (the source-of-truth first)
- ⬜ **Phase 3:** Customer service
- ⬜ **Phase 4:** Product service
- ⬜ **Phase 5:** Account service (introduces cross-service gRPC validation)
- ⬜ **Phase 6:** Transaction orchestrator (the saga layer)
- ⬜ **Phase 7:** Gateway, auth (JWT), Postgres-everywhere, CI, full docker-compose

**Tier 1** (Phases 0–3) — proto contracts + Ledger + Customer working
end-to-end — is the milestone where the project becomes meaningfully
demoable. The remaining phases extend toward a complete system.

---

## Related projects

**[PayFlow](../payflow)** — a smaller, self-contained Go service showcasing a
concurrent payment processor: bounded worker pool, `select`-based shutdown,
race-tested. Built as a focused companion piece to demonstrate the
concurrency patterns that underpin CoreBank's ledger.

---

## About this project (honest section)

I'm a backend engineer with ~3 years of production Go experience in fintech,
mostly building services like the ones modeled here — ledgers, payment
processing, card management. CoreBank is my attempt to design that kind of
system from first principles, in the open, with the reasoning made explicit
for anyone who wants to walk through it.

If you're hiring for Go / backend / fintech / platform roles and any of the
above caught your interest — I'd genuinely love to chat.

📧 rv.visvesh@gmail.com · 💼 [LinkedIn](https://linkedin.com/in/visvesh-ramesh-233254208)

---

## License

MIT — see [LICENSE](LICENSE).
