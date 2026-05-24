# CLAUDE.md — CoreBank

## What this is
A proto-first, multi-module core-banking platform in Go using the Kratos framework.
The contract layer (protos) is complete; implementation is in progress.
Full design reference lives in `docs/ARCHITECTURE.md` — read it before making changes.

## Stack
- Go 1.22+
- Kratos v2 (HTTP + gRPC from one proto via google.api.http annotations)
- buf for proto linting and code generation
- PostgreSQL (one database per service — no shared tables)
- Protocol Buffers, gRPC

## Modules (services)
- `customer`    — identity & KYC lifecycle (state machine)
- `account`     — account lifecycle, types, cached balance; references customer
- `ledger`      — double-entry, append-only, idempotent; SOURCE OF TRUTH for money
- `transaction` — orchestrator (saga) coordinating account + ledger

## Architecture rules (do not violate)
- Dependencies point ONE way: transaction → account → customer; transaction → ledger.
  Never introduce a backward dependency (e.g. customer must never import account).
- Each service owns its own database. No cross-service DB reads — use gRPC.
- Money is ALWAYS int64 minor units + currency string (common.v1.Money). Never floats.
- The ledger is append-only: no update/delete of entries. Reversals are new transactions.
- Every ledger transaction's entries must sum to zero (double-entry invariant).
- Every money-moving RPC takes an idempotency_key; repeats return the original result.
- Status changes are validated state-machine transitions, not arbitrary updates.

## Repo layout
```
api/            proto contracts (source of truth for all interfaces)
  common/v1     Money, pagination, ErrorReason
  customer/v1   customer.proto (types) + customer_service.proto (rpcs)
  account/v1    "
  ledger/v1     "
  transaction/v1 "
app/            one dir per deployable service (cmd / internal/{service,biz,data,server})
pkg/            shared libs: middleware, money, idempotency
docs/           ARCHITECTURE.md — full per-RPC reference
```

## Kratos service layout (per module, when implementing)
- `internal/service` — gRPC/HTTP handlers; thin, maps proto <-> domain
- `internal/biz`     — business logic, domain rules, state-machine validation
- `internal/data`    — repository implementations (Postgres)
- `internal/server`  — HTTP + gRPC server wiring

## Common commands
- Generate code from protos:  `buf generate`
- Lint protos:                `buf lint`
- Build:                      `go build ./...`
- Test (always with race):    `go test -race ./...`
- Vet:                        `go vet ./...`

## Conventions
- Keep messages (types) and services (rpcs) in separate .proto files per module.
- Everything versioned under v1; breaking changes go in v2, never mutate v1.
- Reuse common.v1 types (Money, PageRequest/PageResponse) — do not redefine them per module.
- Run gofmt; CI fails on unformatted code.

## Current state / next steps
- [x] All proto contracts defined (customer, account, ledger, transaction, common)
- [x] Architecture documentation
- [ ] buf.yaml + buf.gen.yaml setup
- [ ] Generate Go stubs
- [ ] Implement first module end-to-end (suggested: ledger or customer)
- [ ] Shared middleware in pkg/
