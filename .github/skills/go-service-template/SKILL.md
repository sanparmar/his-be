---
name: go-service-template
description: "Multi-step workflow skill for generating a complete HIS Go microservice from bounded context discovery through code generation, compilation validation, and test coverage gate. Use when: scaffolding a new service end-to-end, ensuring all layers are consistent."
---

# Go Service Template Skill

This is a multi-step workflow for generating a complete, production-ready HIS microservice.

## Assets

This skill includes these embedded templates (see `templates/` subdirectory):
- `templates/entity.go.tmpl` — Domain entity with audit fields
- `templates/repository.go.tmpl` — Repository interface (port)
- `templates/service.go.tmpl` — Domain service with business rules
- `templates/events.go.tmpl` — Domain event types
- `templates/command_handler.go.tmpl` — CQRS command handler
- `templates/query_handler.go.tmpl` — CQRS query handler
- `templates/postgres_repository.go.tmpl` — sqlx repository implementation
- `templates/kafka_publisher.go.tmpl` — Kafka event publisher
- `templates/grpc_server.go.tmpl` — gRPC server with auth, RBAC, OTel
- `templates/main.go.tmpl` — Entry point with graceful shutdown
- `templates/config.yaml.tmpl` — Default configuration
- `templates/migration.up.sql.tmpl` — Up migration
- `templates/migration.down.sql.tmpl` — Down migration
- `templates/Dockerfile.tmpl` — Multi-stage Docker build
- `templates/Makefile.tmpl` — Build, test, lint, proto, migrate targets
- `templates/README.md.tmpl` — Service documentation

## Workflow Steps

### Step 1: Bounded Context Discovery
**Goal**: Understand what this service owns before writing any code.

Ask the following and confirm answers before proceeding:
1. What is the service name (snake_case, e.g. `patient`, `appointment`, `billing`)?
2. What is the primary aggregate entity (PascalCase)?
3. What state changes does this service own? (list all mutations)
4. What data does this service read (queries)?
5. Are there any PHI fields? List them explicitly.
6. What other services does this call (gRPC clients needed)?
7. What Kafka events does this service publish? What does it consume?

**Deliverable**: Confirmed bounded context definition document (in chat).

---

### Step 2: Proto Contract Definition
**Goal**: Define the gRPC service contract before any implementation.

Generate `proto/<service>/v1/<service>.proto`:
- All entities as proto messages
- All CRUD + domain-specific RPCs
- Standard pagination pattern for List methods
- Health check RPC
- PHI fields annotated with `// PHI - encrypted at rest`

Run `make proto-gen` and confirm generated code compiles before Step 3.

**Gate**: Proto file exists and generates compilable Go code.

---

### Step 3: Domain Layer
**Goal**: Pure domain model with no external dependencies.

Generate:
- `internal/domain/entity.go` — Entity structs, value objects, domain invariants
- `internal/domain/repository.go` — Repository interfaces (ports only)
- `internal/domain/service.go` — Business rules, validation, domain logic
- `internal/domain/events.go` — Domain event types (pure structs)
- `internal/domain/errors.go` — Sentinel errors (`ErrNotFound`, `ErrAlreadyExists`, etc.)

**Gate**: `go build ./internal/domain/...` with ZERO imports from `infrastructure/`.

---

### Step 4: Application Layer (CQRS)
**Goal**: Thin orchestration handlers that delegate to domain.

Generate for each operation:
- `internal/application/command/<action>_<entity>.go`
- `internal/application/query/get_<entity>.go`
- `internal/application/query/list_<entity>s.go`

Each handler receives the command/query, validates, calls domain service, publishes event.

**Gate**: `go build ./internal/application/...` — no domain logic here, only orchestration.

---

### Step 5: Infrastructure Layer
**Goal**: Implement all ports defined in the domain layer.

Generate:
- `internal/infrastructure/postgres/<entity>_repository.go` — sqlx implementation
- `internal/infrastructure/kafka/event_publisher.go` — Kafka producer
- `internal/infrastructure/grpc/server.go` — gRPC handlers with JWT, RBAC, OTel, audit

**Gate**: Each infrastructure file implements the corresponding domain interface exactly.

---

### Step 6: Entry Point & Config
**Goal**: Wire all dependencies and start the server.

Generate:
- `cmd/server/main.go` — DI wiring, OTel init, DB pool, Kafka writer, gRPC server, graceful shutdown
- `configs/config.yaml` — Default config (no secrets)

**Gate**: `go build ./cmd/server/` compiles cleanly.

---

### Step 7: Database Migrations
**Goal**: Idempotent, reversible schema for this bounded context only.

Generate:
- `migrations/000001_create_<entity>s_table.up.sql`
- `migrations/000001_create_<entity>s_table.down.sql`

**Gate**: Migrations run successfully against a local test DB: `make migrate-up`.

---

### Step 8: Tests
**Goal**: ≥80% coverage on domain and application layers.

Generate:
- `internal/domain/service_test.go` — Table-driven, mocked repository
- `internal/application/command/<action>_<entity>_test.go`
- `internal/infrastructure/grpc/server_test.go`

**Gate**: `go test ./... -race -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total` shows ≥80%.

---

### Step 9: Build Artifacts
**Goal**: Containerize and document.

Generate:
- `Dockerfile` — Multi-stage distroless build
- `Makefile` — All standard targets
- `README.md` — Setup, env vars, running, testing, proto regen

**Gate**: `docker build -t his-<service>:local .` succeeds.

---

### Step 10: Final Validation
Run the complete checklist:

```bash
go build ./...         # must pass
go test ./... -race    # must pass, ≥80% coverage
golangci-lint run      # must pass (0 errors)
```

Report:
- [ ] All files generated
- [ ] Compilation: PASS
- [ ] Tests: PASS (coverage: __)
- [ ] Lint: PASS
- [ ] Kafka topics: listed
- [ ] gRPC service port: configured
- [ ] PHI fields: encrypted, not logged
- [ ] Migration: up + down verified
