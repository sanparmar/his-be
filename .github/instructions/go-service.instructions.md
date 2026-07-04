---
applyTo: "services/**/*.go"
---

# Go Service File Conventions

## Package Naming
- `package domain` for all files in `internal/domain/`
- `package command` / `package query` for application layer files
- `package postgres` / `package kafka` / `package grpc` for infrastructure files

## Entity Definition Pattern
Every domain entity MUST follow this pattern:

```go
type <Entity> struct {
    ID        uuid.UUID  `json:"id"         db:"id"          validate:"required,uuid4"`
    TenantID  uuid.UUID  `json:"tenant_id"  db:"tenant_id"   validate:"required,uuid4"`
    // ... domain fields ...
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
    CreatedBy uuid.UUID  `json:"created_by" db:"created_by"  validate:"required,uuid4"`
    UpdatedBy uuid.UUID  `json:"updated_by" db:"updated_by"  validate:"required,uuid4"`
    DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}
```

## Repository Interface Pattern
```go
type <Entity>Repository interface {
    Create(ctx context.Context, entity *<Entity>) error
    GetByID(ctx context.Context, tenantID, id uuid.UUID) (*<Entity>, error)
    Update(ctx context.Context, entity *<Entity>) error
    Delete(ctx context.Context, tenantID, id uuid.UUID) error
    List(ctx context.Context, tenantID uuid.UUID, filters <Entity>Filters) ([]*<Entity>, int64, error)
}
```

## Command Handler Pattern
```go
type <Action><Entity>Handler struct {
    repo   <Entity>Repository
    events EventPublisher
    logger zerolog.Logger
    tracer trace.Tracer
}

func (h *<Action><Entity>Handler) Handle(ctx context.Context, cmd <Action><Entity>Command) error {
    ctx, span := h.tracer.Start(ctx, "<Domain>.<Action><Entity>")
    defer span.End()
    span.SetAttributes(attribute.String("tenant.id", cmd.TenantID.String()))
    // ... implementation
}
```

## Error Wrapping
Always wrap errors with operation context:
```go
if err != nil {
    return fmt.Errorf("<handlerName>: %w", err)
}
```

## PHI Field Handling
- Never log PHI (patient name, DOB, MRN, diagnosis, medications)
- Use field-level encryption via `libs/crypto` for PHI at rest
- PHI fields in span attributes: use anonymized reference IDs only

## Context Requirement
Every exported function must accept `ctx context.Context` as its first argument.

## Nil Check Before Pointer Dereference
Always check pointers before dereferencing, especially `DeletedAt *time.Time`.

## Test File Co-location
Place `<file>_test.go` in the same directory as the file it tests.
Use table-driven tests with `t.Run(tc.name, func(t *testing.T) {...})`.

## S3/MinIO Object Storage Pattern

For services handling file uploads/downloads (DICOM images, PDFs, documents):

### Never Stream Files Through gRPC
**DON'T**: Send file bytes via gRPC server streaming
```go
// ❌ WRONG: This causes memory bloat and gRPC message size limits
func (h *Handler) DownloadDocument(req *pb.DownloadRequest, stream pb.DocumentService_DownloadServer) error {
    data, _ := h.storage.Read(ctx)
    stream.Send(&pb.Chunk{Data: data})
}
```

**DO**: Use presigned URLs for direct client-to-S3 access
```go
// ✓ CORRECT: gRPC returns URL, client downloads directly from S3
func (h *Handler) GetDownloadUrl(ctx context.Context, req *pb.GetDownloadUrlRequest) (*pb.GetDownloadUrlResponse, error) {
    url, _ := h.s3Client.GetPresignedDownloadURL(ctx, objectKey)
    return &pb.GetDownloadUrlResponse{DownloadUrl: url}, nil
}
```

### S3 Client Configuration
- Use `github.com/aws/aws-sdk-go-v2/service/s3` (works with both AWS S3 and MinIO)
- Configure endpoint from environment: `S3_PROVIDER` + `S3_ENDPOINT`
- MinIO requires `UsePathStyle: true` option
- Always set presigned URL expiry: 10 min for upload, 1 hour for download

### Object Key Naming
Prefix object keys with tenant ID to ensure multi-tenant isolation:
```go
objectKey := fmt.Sprintf("%s/documents/%s/%s", tenantID, docID, fileName)
// Result: "550e8400-e29b-41d4-a716-446655440000/documents/abc123/report.pdf"
```

### Presigned URL Expiry
- **Upload URLs**: 10 minutes (quick transaction)
- **Download URLs**: 1 hour (user may be slow)
- Never create permanent URLs (security risk)

### Audit Logging
Every file upload/download must be logged:
```go
h.auditLog(ctx, "document.uploaded", tenantID, docID, fileName, uploadedBy)
h.auditLog(ctx, "document.downloaded", tenantID, docID, fileName, downloadedBy)
```

### Error Handling
```go
if err != nil {
    span.RecordError(err)
    // Use specific gRPC status codes
    return nil, status.Error(codes.Internal, "failed to generate presigned URL")
}
```

### Testing
- **Unit tests**: Mock the S3Client interface
- **Integration tests**: Use MinIO testcontainer
```go
import "github.com/testcontainers/testcontainers-go"

minioContainer, _ := testcontainers.GenericContainer(ctx, ...)
endpoint := minioContainer.Host(ctx) + ":9000"
s3Client := NewS3Client(S3Config{Provider: "minio", Endpoint: endpoint})
```

### Quota Validation
Always validate file size before accepting upload:
```go
const MaxFileSizeBytes = 500 * 1024 * 1024 // 500 MB

if req.FileSizeBytes > MaxFileSizeBytes {
    return nil, status.Error(codes.InvalidArgument, "file too large")
}
```

See [ADR 0001: S3/MinIO Strategy](../../../adr/0001-s3-minio-cloud-onprem-storage.md) for architectural rationale.
