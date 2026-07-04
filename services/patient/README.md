# Patient Service

A production-ready Go microservice for the Patient bounded context in the Hospital Information System (HIS).

## Overview

The Patient service handles:
- **Registration**: Register new patients with auto-generated MRN
- **Retrieval**: Get patient details with tenant isolation
- **Search**: Full-text search by name, MRN, phone, email
- **Update**: Modify patient contact and personal information
- **Activation**: Activate registered patients for clinical use

All PHI fields (DOB, phone, email) are encrypted at rest in PostgreSQL. All mutations are audit-logged and trigger Kafka domain events.

## Architecture

```
cmd/server/main.go          ← Entry point (DI wiring, graceful shutdown)
│
├── internal/
│   ├── domain/              ← Pure business logic, no framework imports
│   │   ├── entity.go        ← Patient aggregate root
│   │   ├── service.go       ← Domain business rules
│   │   ├── repository.go    ← Repository interface (port)
│   │   ├── events.go        ← Domain events
│   │   └── errors.go        ← Sentinel errors
│   │
│   ├── application/         ← CQRS handlers
│   │   ├── command/         ← Commands (CreatePatient, UpdatePatient, ActivatePatient)
│   │   └── query/           ← Queries (GetPatient, SearchPatients, ListPatients)
│   │
│   └── infrastructure/      ← Implementations
│       ├── postgres/        ← Repository (adapter)
│       ├── kafka/           ← Event publisher (adapter)
│       └── grpc/            ← gRPC server (handler)
│
├── pkg/
│   ├── config/              ← Configuration management (12-factor)
│   └── logger/              ← Structured logging with PHI masking
│
├── migrations/              ← SQL migrations (up/down)
├── configs/                 ← Default configuration (no secrets)
├── proto/                   ← gRPC service definitions
└── Dockerfile               ← Multi-stage distroless build
```

## Tech Stack

- **Language**: Go 1.22+
- **Database**: PostgreSQL 13+
- **Message Queue**: Apache Kafka 3.0+
- **RPC**: gRPC + Protocol Buffers
- **Logging**: zerolog (structured JSON)
- **Testing**: testify, sqlc, testcontainers-go
- **Deployment**: Docker, Kubernetes

## Prerequisites

- Go 1.22+
- PostgreSQL 13+
- Apache Kafka 3.0+
- Docker & Docker Compose (for local development)
- `protoc` compiler + Go plugins (for proto code generation)
- `migrate` tool (for database migrations)

## Getting Started

### 1. Clone and Setup

```bash
cd services/patient
go mod download
make build
```

### 2. Local Development (Docker Compose)

Start PostgreSQL, Redis, and Kafka:

```bash
docker-compose up -d
```

### 3. Run Database Migrations

```bash
make migrate-up
```

### 4. Start the Service

```bash
make dev
# Or manually:
DB_HOST=localhost DB_USER=postgres DB_PASSWORD=postgres DB_NAME=his_patient \
KAFKA_BROKERS=localhost:9092 \
./bin/patient-service
```

The service listens on:
- **gRPC**: `:50051`
- **Health Check**: `:8081` (`GET /health`)
- **Metrics**: `:8081` (`GET /metrics`)

### 5. Run Tests

```bash
make test           # Unit tests
make cover          # Coverage report
make test-integration  # Integration tests with testcontainers
```

## Configuration

Configuration is loaded from (in order):
1. `configs/config.yaml` (defaults)
2. Environment variables (override)

### Environment Variables

```bash
GRPC_PORT=50051                    # gRPC server port
DB_HOST=localhost                  # PostgreSQL host
DB_PORT=5432                       # PostgreSQL port
DB_USER=postgres                   # PostgreSQL user
DB_PASSWORD=postgres               # PostgreSQL password
DB_NAME=his_patient                # PostgreSQL database name
KAFKA_BROKERS=localhost:9092       # Kafka broker addresses (comma-separated)
KAFKA_GROUP_ID=his-patient-service # Kafka consumer group
LOG_LEVEL=info                     # Logging level (debug, info, warn, error)
```

## gRPC API

See [proto/patient/v1/patient.proto](proto/patient/v1/patient.proto) for service definitions.

### Core RPCs

**CreatePatient** — Register a new patient
```protobuf
rpc CreatePatient(CreatePatientRequest) returns (CreatePatientResponse);
```

**GetPatient** — Retrieve a single patient
```protobuf
rpc GetPatient(GetPatientRequest) returns (GetPatientResponse);
```

**SearchPatients** — Search patients by criteria
```protobuf
rpc SearchPatients(SearchPatientsRequest) returns (SearchPatientsResponse);
```

**ListPatients** — List all patients with pagination
```protobuf
rpc ListPatients(ListPatientsRequest) returns (ListPatientsResponse);
```

**UpdatePatient** — Update patient contact/personal info
```protobuf
rpc UpdatePatient(UpdatePatientRequest) returns (UpdatePatientResponse);
```

**ActivatePatient** — Mark patient as ACTIVE
```protobuf
rpc ActivatePatient(ActivatePatientRequest) returns (ActivatePatientResponse);
```

**Health** — Service health check
```protobuf
rpc Health(google.protobuf.Empty) returns (HealthResponse);
```

### Authentication & Authorization (RBAC)

Every RPC requires a valid JWT token in the `authorization` header. User and Tenant IDs are extracted from JWT claims via gRPC metadata:

```bash
Authorization: Bearer eyJhbGc...
X-User-ID: <user-uuid>
X-Tenant-ID: <tenant-uuid>
```

Required roles per RPC:
- `CreatePatient`: ADMIN, CLERK
- `GetPatient`: ADMIN, DOCTOR, NURSE, CLERK
- `SearchPatients`: ADMIN, DOCTOR, NURSE, CLERK
- `ListPatients`: ADMIN, DOCTOR, NURSE
- `UpdatePatient`: ADMIN, CLERK
- `ActivatePatient`: ADMIN, CLERK

## Domain Events (Kafka)

This service publishes events to Kafka:

| Event | Topic | Trigger | Schema |
|-------|-------|---------|--------|
| PatientRegistered | `his.patient.patient.registered` | CreatePatient RPC | { patient_id, mrn, first_name, ... } |
| PatientUpdated | `his.patient.patient.updated` | UpdatePatient RPC | { patient_id, mrn, updated_fields, ... } |
| PatientActivated | `his.patient.patient.activated` | ActivatePatient RPC | { patient_id, mrn, status, ... } |

Event payloads are JSON and include:
- `aggregate_id` (patient UUID)
- `tenant_id` (hospital UUID)
- `timestamp` (event time)
- `user_id` (who triggered the change)

### Consuming Events

Other services can subscribe to these topics to:
- Update search indexes (Elasticsearch)
- Trigger workflows (approval, notifications)
- Sync caches (Redis)
- Generate audit logs (separate audit service)

## Database Schema

### patients table

```sql
CREATE TABLE patients (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,           -- Which hospital
    mrn VARCHAR(50) NOT NULL UNIQUE,   -- Medical Record Number
    first_name VARCHAR(255),           -- PHI encrypted
    last_name VARCHAR(255),            -- PHI encrypted
    dob DATE,                          -- PHI encrypted
    gender VARCHAR(20),
    phone VARCHAR(20),                 -- PHI encrypted
    email VARCHAR(255),                -- PHI encrypted
    status VARCHAR(50),                -- REGISTERED, ACTIVE, INACTIVE, DISCHARGED
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMP               -- Soft delete
);
```

Indexes:
- `(tenant_id)` — Tenant isolation
- `(mrn, tenant_id)` — MRN lookup
- `(status)` — Status filtering
- `(created_at DESC)` — Recent first
- `(first_name)`, `(last_name)`, `(phone)`, `(email)` — Search

### patient_audit_logs table

Tracks all mutations for compliance (HIPAA, GDPR, DPDP).

## Security Considerations

### PHI Protection

- **At Rest**: DOB, phone, email encrypted in PostgreSQL (via PgCrypto plugin)
- **In Transit**: TLS 1.3 for all gRPC and Kafka connections
- **In Logs**: PHI fields masked in structured logs (see `pkg/logger/`)
- **In Memory**: No PHI in Redis or caches; only UUIDs

### Multi-Tenancy

Every query includes `tenant_id` check to prevent cross-tenant data access:

```go
// GetByID enforces tenant isolation
GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Patient, error)
```

### Audit Logging

All mutations are logged to `patient_audit_logs` table:
- WHO (user_id)
- WHAT (action: CREATE, UPDATE, DELETE)
- WHEN (timestamp)
- WHERE (patient_id)
- FIELDS_CHANGED (JSONB delta)

## Development

### Proto Code Generation

After modifying `.proto` files:

```bash
make proto-gen
```

This generates:
- `api/patient/v1/*.pb.go` (messages)
- `api/patient/v1/*_grpc.pb.go` (service stubs)

### Adding a New Migration

```bash
# Create migration file
touch migrations/000002_add_photo_field.up.sql
touch migrations/000002_add_photo_field.down.sql

# Edit files
# Up: ALTER TABLE patients ADD COLUMN photo_url VARCHAR(500);
# Down: ALTER TABLE patients DROP COLUMN photo_url;

# Run migration
make migrate-up
```

### Running Tests

```bash
# Unit tests (domain + application layers)
make test

# With coverage
make cover

# Integration tests (with Postgres + Kafka containers)
make test-integration

# Watch mode (requires `entr` tool)
make run-tests-watch
```

Coverage targets:
- **Domain**: ≥90% (pure logic)
- **Application**: ≥85% (orchestration)
- **Infrastructure**: ≥70% (adapter code)
- **Overall**: ≥80%

### Linting

```bash
make lint
```

Enforces:
- Go fmt
- Go vet
- Cyclomatic complexity
- Error handling
- Unused variables

## Deployment

### Docker Build

```bash
make docker
```

Produces: `his-patient:latest` (multi-stage, distroless, ~20MB)

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: patient-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: patient-service
  template:
    metadata:
      labels:
        app: patient-service
    spec:
      containers:
      - name: patient-service
        image: his-patient:latest
        ports:
        - containerPort: 50051
        - containerPort: 8081
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: patient-config
              key: db-host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: patient-secrets
              key: db-password
        # ... more env vars
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
```

### Health Checks

The service exposes health checks:

- **HTTP**: `GET :8081/health` → `{"status":"UP","service":"patient","version":"1.0.0"}`
- **gRPC**: `grpc_health_v1.Check()`

## Performance

### Benchmarks

Run benchmarks:

```bash
go test -bench=. -benchmem ./internal/domain/...
go test -bench=. -benchmem ./internal/infrastructure/postgres/...
```

Expected performance:
- Patient registration: <50ms (p95)
- Patient retrieval: <20ms (p95)
- Search (1000 records): <100ms (p95)

### Optimization Tips

1. **Database**: Ensure indexes are used (`EXPLAIN ANALYZE`)
2. **Kafka**: Batch events if throughput is high
3. **Caching**: Add Redis layer for frequently-searched patients (future)
4. **Connection Pooling**: Set `DB_MAX_CONNS` based on load (default: 25)

## Troubleshooting

### Service Won't Start

1. Check PostgreSQL is running: `psql -h localhost -U postgres`
2. Check Kafka is running: `kafka-broker-api-versions.sh --bootstrap-server localhost:9092`
3. Check port availability: `lsof -i :50051`

### High Latency

1. Check database indexes: `SELECT * FROM pg_stat_user_indexes WHERE idx_scan = 0;`
2. Check Kafka lag: `kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group his-patient-service`
3. Check gRPC tracer logs: `LOG_LEVEL=debug`

### Data Not Appearing

1. Check Kafka topics exist: `kafka-topics.sh --bootstrap-server localhost:9092 --list`
2. Check migrations ran: `SELECT * FROM schema_migrations;`
3. Check audit logs: `SELECT * FROM patient_audit_logs ORDER BY changed_at DESC LIMIT 10;`

## Contributing

1. Follow Go best practices (effective Go, code review)
2. Write failing tests first (TDD)
3. Ensure ≥80% coverage
4. Run `make lint` before committing
5. Submit PR with clear description

## License

Proprietary — Deloitte US Consulting LLP

## Contacts

- **Service Owner**: HIS Platform Team
- **Architecture**: Platform Architecture Group
- **Support**: #his-patient-service Slack channel
