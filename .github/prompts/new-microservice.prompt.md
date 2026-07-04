---
name: new-microservice
description: "Scaffold a complete new HIS microservice with Clean Architecture, DDD, gRPC, Kafka, and Postgres. Use when: creating a new bounded context service, generating a new domain service from scratch."
---

# New Microservice Scaffold

Generate a complete HIS microservice for the following bounded context.

## Inputs

- **Service name**: ${input:serviceName:Enter the service domain name (e.g. patient, appointment, billing)}
- **Entity name**: ${input:entityName:Enter the primary domain entity name (e.g. Patient, Appointment, Invoice)}
- **Domain description**: ${input:domainDescription:Describe what this service owns and is responsible for}

## Instructions

Using the inputs above, generate ALL of the following files. Do not skip any. Generate production-quality code — no placeholder comments.

### 1. Proto Definition
File: `proto/${serviceName}/v1/${serviceName}.proto`

- Define the `${entityName}` message with all domain fields
- Define `${serviceName^}Service` with CRUD + List methods
- Include health check RPC
- PHI fields marked with `// PHI - encrypted at rest`

### 2. Domain Layer
Files: `services/${serviceName}/internal/domain/`

- `entity.go` — `${entityName}` struct with audit fields, TenantID, and validate tags
- `repository.go` — `${entityName}Repository` interface with CRUD + List + soft-delete
- `service.go` — `${entityName}Service` with business rules, calls repository
- `events.go` — Domain event types: `${entityName}Created`, `${entityName}Updated`, `${entityName}Deleted`
- `errors.go` — Sentinel errors: `ErrNotFound`, `ErrAlreadyExists`, `ErrForbidden`

### 3. Application Layer (CQRS)
Files: `services/${serviceName}/internal/application/`

- `command/create_${entityName}.go` — `Create${entityName}Command` struct + `Create${entityName}Handler`
- `command/update_${entityName}.go` — `Update${entityName}Command` + handler
- `command/delete_${entityName}.go` — `Delete${entityName}Command` + handler
- `query/get_${entityName}.go` — `Get${entityName}Query` + handler
- `query/list_${entityName}s.go` — `List${entityName}sQuery` + handler

### 4. Infrastructure Layer
Files: `services/${serviceName}/internal/infrastructure/`

- `postgres/${entityName}_repository.go` — Full sqlx implementation of `${entityName}Repository`
- `kafka/event_publisher.go` — Kafka producer publishing `his.${serviceName}.*` events
- `grpc/server.go` — gRPC server implementing all proto methods with JWT extraction, RBAC check, OTel spans, and audit logging

### 5. Entry Point
File: `services/${serviceName}/cmd/server/main.go`

- Wire dependencies with `google.golang.org/wire` or manual DI
- Parse config via Viper from env + `configs/config.yaml`
- Start gRPC server with graceful shutdown on SIGTERM/SIGINT
- Initialize OpenTelemetry, zerolog, and database connection pool

### 6. Config
File: `services/${serviceName}/configs/config.yaml`

```yaml
server:
  port: 50051
  grpc_timeout: 30s

database:
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

kafka:
  brokers: ["localhost:9092"]
  topic_prefix: "his.${serviceName}"

otel:
  service_name: "his-${serviceName}"
  exporter: "otlp"
```

### 7. Database Migration
Files: `services/${serviceName}/migrations/`

- `000001_create_${entityName}s_table.up.sql` — Full table with audit columns, indexes
- `000001_create_${entityName}s_table.down.sql` — Reversible drop

### 8. Dockerfile
File: `services/${serviceName}/Dockerfile`

Multi-stage build:
- Stage 1: `golang:1.22-alpine` builder
- Stage 2: `gcr.io/distroless/static-debian12` runtime
- Non-root user, no shell

### 9. Makefile
File: `services/${serviceName}/Makefile`

Targets: `build`, `test`, `lint`, `proto-gen`, `migrate-up`, `migrate-down`, `docker-build`, `run`

### 10. README
File: `services/${serviceName}/README.md`

Include: purpose, setup, env vars, running locally, running tests, proto regeneration.

### 11. Unit Tests
Files co-located with each file:

- `services/${serviceName}/internal/domain/service_test.go` — Table-driven tests, mock repository
- `services/${serviceName}/internal/infrastructure/grpc/server_test.go` — gRPC handler tests with mock application layer

## Quality Gates

After generating:
- All files must compile: `go build ./...`
- Tests pass: `go test ./... -race`
- Linter passes: `golangci-lint run`
- ≥80% test coverage on domain and application layers
