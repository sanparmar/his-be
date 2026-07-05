# Appointment Service

Phase 1 appointment backend slice for prototype scheduling in HIS.

## Scope

This scaffold follows the existing patient service layering while staying intentionally lightweight:

- `internal/domain` — appointment aggregate, invariants, repository port
- `internal/application` — create/update-status commands and list/search queries
- `internal/infrastructure/memory` — in-memory repository for prototype flows and tests
- `internal/infrastructure/http` — standard-library JSON adapter for prototype integration
- `pkg/config` — env-driven configuration with safe defaults
- `migrations` — PostgreSQL-ready schema for later persistence wiring

## Aggregate

Each appointment is tenant-scoped and stores:

- appointment id
- patient MRN and summary
- doctor and department
- appointment date and slot time
- visit type
- status
- notes
- created/updated audit fields

## HTTP Endpoints

- `GET /health`
- `POST /appointments`
- `GET /appointments`
- `GET /appointments/search?q=...`
- `PATCH /appointments/{id}/status`

Multi-tenancy is passed with `X-Tenant-ID` header for read/update endpoints. Create accepts `tenant_id` in the request body.

## Run

```bash
cd services/appointment
go test ./...
go run ./cmd/server
```

Default server address: `:8082`
