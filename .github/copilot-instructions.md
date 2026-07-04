# HIS Backend — GitHub Copilot Instructions

You are a Principal Go Engineer and Solution Architect for an enterprise Hospital Information System (HIS).
This repo contains Go microservices following Domain-Driven Design and Clean Architecture.

## Architecture Constraints (Non-Negotiable)

- **Database per bounded context.** No service may query or import another service's database schema.
- **Synchronous communication via gRPC only.** No direct HTTP calls between services.
- **Asynchronous communication via Kafka.** Use domain events for cross-service side effects.
- **Every service publishes domain events** on state changes.
- **JWT + RBAC** on every gRPC handler. Never skip auth.
- **Audit log** every mutation (who, what, when, tenant, correlation ID).

## Service Structure (Mandatory)

Every microservice MUST follow this layout exactly:

```
services/<domain>/
  cmd/
    server/
      main.go          # wire everything, graceful shutdown
  internal/
    domain/
      entity.go        # pure domain types, no framework imports
      repository.go    # repository interface (port)
      service.go       # domain service logic
      events.go        # domain event types
    application/
      command/         # CQRS commands + handlers
      query/           # CQRS queries + handlers
    infrastructure/
      postgres/        # repository implementations
      kafka/           # event publisher implementations
      grpc/            # gRPC server handlers
  pkg/
    config/            # env-based config with validation
    logger/            # structured zerolog wrapper
    telemetry/         # OpenTelemetry setup
  api/
    v1/                # generated gRPC + REST gateway code (DO NOT EDIT)
  proto/               # .proto source files for this service
  migrations/          # numbered SQL migrations (up/down)
  configs/
    config.yaml        # default config (no secrets)
  Dockerfile
  Makefile
  README.md
```

## Go Coding Standards

- **Go 1.22+**. Use `errors.Is` / `errors.As`; never compare error strings.
- **Struct tags**: always include `json`, `db`, and `validate` tags on domain structs.
- **Context propagation**: first argument of every function must be `ctx context.Context`.
- **Table-driven tests** with `t.Run` for all unit tests.
- **Mocks**: use `mockery` generated interfaces; never mock concrete types.
- **No globals** except logger and tracer initialized in `main.go`.
- **Configuration**: 12-factor — all config via env vars with a `config.yaml` default fallback.
- **Error wrapping**: `fmt.Errorf("operationName: %w", err)` — always wrap with context.
- **Linting**: code must pass `golangci-lint` with the repo's `.golangci.yml` config.

## Healthcare & Compliance Rules

- **Multi-tenancy**: every domain entity must carry `TenantID uuid.UUID`.
- **PHI fields** (patient name, DOB, MRN, diagnosis, medications): always encrypted at rest; never log raw PHI.
- **FHIR R4 alignment**: domain models must map cleanly to FHIR R4 resource types.
- **HL7 v2 parsing**: use the `hapi` or equivalent library; never hand-parse HL7 segments.
- **Audit fields**: every mutable entity needs `CreatedAt`, `UpdatedAt`, `CreatedBy`, `UpdatedBy`, `DeletedAt`.
- **HIPAA**: no PHI in logs, traces, or metrics labels.

## OpenTelemetry (Required on Every Handler)

```go
ctx, span := tracer.Start(ctx, "ServiceName.OperationName")
defer span.End()
span.SetAttributes(attribute.String("tenant.id", tenantID))
// record errors:
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
```

## Testing Requirements

- **≥80% coverage** on all new/changed code.
- Unit tests: pure, no I/O, mock all dependencies.
- Integration tests: use `testcontainers-go` for Postgres and Kafka.
- Every test file ends in `_test.go` and lives alongside the file it tests.

## Kafka Event Naming Convention

```
his.<domain>.<entity>.<past-tense-verb>
# Examples:
his.patient.patient.registered
his.appointment.slot.booked
his.billing.invoice.finalized
```

## gRPC Naming Convention

- Package: `his.<domain>.v1`
- Service: `<Domain>Service`
- Methods: `Create<Entity>`, `Get<Entity>`, `Update<Entity>`, `Delete<Entity>`, `List<Entity>s`
- Request/Response: `<Method>Request` / `<Method>Response`

## Dependencies (Preferred Libraries)

| Concern | Library |
|---------|---------|
| gRPC | `google.golang.org/grpc` |
| Kafka | `github.com/segmentio/kafka-go` |
| DB | `github.com/jmoiron/sqlx` + `pgx/v5` |
| Migrations | `github.com/golang-migrate/migrate/v4` |
| Config | `github.com/spf13/viper` |
| Logging | `github.com/rs/zerolog` |
| Validation | `github.com/go-playground/validator/v10` |
| Tracing | `go.opentelemetry.io/otel` |
| DI | `google.golang.org/wire` |
| Mocking | `github.com/stretchr/testify/mock` + `mockery` |
| Containers | `github.com/testcontainers/testcontainers-go` |
| UUID | `github.com/google/uuid` |
| Object Storage | `github.com/aws/aws-sdk-go-v2/service/s3` (S3 + MinIO) |

## Object Storage (S3 / MinIO)

For **file uploads/downloads** (DICOM images, PDFs, scanned documents):

- Use **AWS S3 SDK v2** which works with both AWS S3 and MinIO (S3-compatible)
- **Never stream files through gRPC** — use presigned URLs instead
- Architecture: Client gets presigned URL via gRPC RPC → Client uploads/downloads directly to S3/MinIO
- **Cloud deployments**: Use AWS S3
- **On-premise deployments**: Use MinIO (self-hosted, open-source)
- See [ADR 0001: S3/MinIO Cloud & On-Prem Storage](../../../adr/0001-s3-minio-cloud-onprem-storage.md) for full strategy
- Development: Run MinIO in Docker container (`docker-compose up minio`)
