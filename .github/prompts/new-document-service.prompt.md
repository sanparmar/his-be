---
name: new-document-service
description: "Scaffold a complete HIS document service with S3/MinIO presigned URLs, audit logging, and lifecycle management. Use when: building file upload/download features, implementing medical document storage."
---

# Create a New Document Service

Scaffold a document service with S3/MinIO integration, presigned URLs, and audit logging.

## Inputs

- **Domain**: ${input:domain:Which domain? (e.g. documents, records, imaging, lab)}
- **Document types**: ${input:types:What document types? (e.g. discharge_summary, lab_report, imaging_study)}
- **Max file size**: ${input:maxsize:Max file size in MB? (default: 500)}

## Service Overview

This service handles:
- **Presigned URL generation** for secure uploads/downloads
- **Multi-tenant document isolation** (tenant_id required)
- **Audit logging** of all uploads/downloads (who, what, when)
- **Document metadata** (name, type, size, created_by)
- **Lifecycle management** (archive, delete)
- **S3/MinIO compatible** (cloud and on-premise)

## Generated Files

### 1. Proto Definition
**File**: `proto/his/documents/v1/document.proto`

```protobuf
syntax = "proto3";

package his.documents.v1;

import "google/protobuf/timestamp.proto";

service DocumentService {
  rpc InitiateUpload(InitiateUploadRequest) returns (InitiateUploadResponse);
  rpc CompleteUpload(CompleteUploadRequest) returns (CompleteUploadResponse);
  rpc GetDownloadUrl(GetDownloadUrlRequest) returns (GetDownloadUrlResponse);
  rpc GetDocument(GetDocumentRequest) returns (GetDocumentResponse);
  rpc ListDocuments(ListDocumentsRequest) returns (ListDocumentsResponse);
  rpc DeleteDocument(DeleteDocumentRequest) returns (DeleteDocumentResponse);
}

message InitiateUploadRequest {
  string tenant_id = 1; // required: multi-tenant isolation
  string patient_id = 2; // required or optional depending on domain
  DocumentType document_type = 3; // required: enum
  string file_name = 4; // required: original filename
  int64 file_size_bytes = 5; // required: for quota validation
  string content_type = 6; // required: e.g. application/pdf, image/dicom
}

message InitiateUploadResponse {
  string upload_session_id = 1; // UUID for tracking
  string upload_url = 2; // presigned S3 URL (10 min expiry)
  map<string, string> headers = 3; // additional headers to send with PUT
}

message CompleteUploadRequest {
  string tenant_id = 1;
  string upload_session_id = 2;
  string s3_etag = 3; // ETag returned by S3 after PUT
}

message CompleteUploadResponse {
  string document_id = 1; // newly created document ID
  Document metadata = 2;
}

message GetDownloadUrlRequest {
  string tenant_id = 1;
  string document_id = 2;
}

message GetDownloadUrlResponse {
  string download_url = 1; // presigned S3 URL (1 hour expiry)
  Document metadata = 2;
}

message GetDocumentRequest {
  string tenant_id = 1;
  string document_id = 2;
}

message GetDocumentResponse {
  Document document = 1;
}

message ListDocumentsRequest {
  string tenant_id = 1;
  string patient_id = 2; // optional: filter by patient
  DocumentType document_type = 3; // optional: filter by type
  int32 limit = 4; // default 20, max 100
  int32 offset = 5;
}

message ListDocumentsResponse {
  repeated Document documents = 1;
  int32 total_count = 2;
}

message DeleteDocumentRequest {
  string tenant_id = 1;
  string document_id = 2;
  string reason = 3; // optional: deletion reason for audit
}

message DeleteDocumentResponse {}

message Document {
  string id = 1; // primary key (UUID)
  string tenant_id = 2;
  string patient_id = 3;
  DocumentType document_type = 4;
  string file_name = 5;
  int64 file_size_bytes = 6;
  string content_type = 7;
  string s3_object_key = 8; // internal S3 path
  DocumentStatus status = 9; // uploaded, archived, deleted
  google.protobuf.Timestamp created_at = 10;
  string created_by = 11; // audit: who uploaded
  google.protobuf.Timestamp updated_at = 12;
  string updated_by = 13;
  google.protobuf.Timestamp deleted_at = 14;
}

enum DocumentType {
  TYPE_UNSPECIFIED = 0;
  DISCHARGE_SUMMARY = 1;
  OPERATIVE_NOTE = 2;
  LAB_REPORT = 3;
  IMAGING_STUDY = 4;
  PATHOLOGY_REPORT = 5;
  PRESCRIPTION = 6;
  CONSENT_FORM = 7;
  MEDICAL_RECORD = 8;
  OTHER = 9;
}

enum DocumentStatus {
  STATUS_UNSPECIFIED = 0;
  UPLOADED = 1; // successfully uploaded to S3
  ARCHIVED = 2; // moved to cold storage
  DELETED = 3;
}

// Domain event: document uploaded
message DocumentUploadedEvent {
  string event_type = 1; // "his.documents.document.uploaded"
  string event_id = 2;
  google.protobuf.Timestamp occurred_at = 3;
  string tenant_id = 4;
  string document_id = 5;
  string patient_id = 6;
  DocumentType document_type = 7;
  string file_name = 8;
  int64 file_size_bytes = 9;
  string uploaded_by = 10;
}
```

### 2. Entity & Repository
**File**: `services/documents/internal/domain/entity.go`

```go
package domain

import (
  "time"
  "github.com/google/uuid"
)

// Document represents a medical document
type Document struct {
  ID           uuid.UUID
  TenantID     uuid.UUID
  PatientID    uuid.UUID
  Type         DocumentType
  FileName     string
  FileSizeBytes int64
  ContentType  string
  S3ObjectKey  string
  Status       DocumentStatus
  
  CreatedAt time.Time
  CreatedBy uuid.UUID
  UpdatedAt time.Time
  UpdatedBy uuid.UUID
  DeletedAt *time.Time
}

type DocumentType string

const (
  DocumentTypeDischargeSummary = "discharge_summary"
  DocumentTypeOperativeNote    = "operative_note"
  DocumentTypeLabReport        = "lab_report"
  DocumentTypeImagingStudy     = "imaging_study"
)

type DocumentStatus string

const (
  DocumentStatusUploaded  = "uploaded"
  DocumentStatusArchived  = "archived"
  DocumentStatusDeleted   = "deleted"
)

// DocumentRepository is the port for document persistence
type DocumentRepository interface {
  Create(ctx context.Context, doc *Document) error
  GetByID(ctx context.Context, tenantID, docID uuid.UUID) (*Document, error)
  ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID, limit, offset int) ([]*Document, int, error)
  Update(ctx context.Context, doc *Document) error
  Delete(ctx context.Context, tenantID, docID uuid.UUID) error
}
```

### 3. S3 Client Wrapper
**File**: `services/documents/internal/infrastructure/storage/s3_client.go`

```go
package storage

import (
  "context"
  "fmt"
  "time"
  
  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/credentials"
  "github.com/aws/aws-sdk-go-v2/service/s3"
  "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Config struct {
  Provider  string // "aws" or "minio"
  Endpoint  string // only for MinIO
  Bucket    string
  Region    string
  AccessKey string
  SecretKey string
}

type S3Client struct {
  client   *s3.Client
  bucket   string
  provider string
}

func NewS3Client(cfg S3Config) (*S3Client, error) {
  var awsCfg aws.Config
  var err error
  
  if cfg.Provider == "minio" {
    // MinIO requires credentials
    awsCfg, err = config.LoadDefaultConfig(context.Background(),
      config.WithRegion(cfg.Region),
      config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        cfg.AccessKey, cfg.SecretKey, "",
      )),
    )
    if err != nil {
      return nil, fmt.Errorf("minio config: %w", err)
    }
    
    return &S3Client{
      client: s3.NewFromConfig(awsCfg, func(o *s3.Options) {
        o.UsePathStyle = true // MinIO requires path-style URLs
        o.BaseEndpoint = cfg.Endpoint
      }),
      bucket:   cfg.Bucket,
      provider: cfg.Provider,
    }, nil
  }
  
  // AWS S3 (default)
  awsCfg, err = config.LoadDefaultConfig(context.Background())
  if err != nil {
    return nil, fmt.Errorf("aws s3 config: %w", err)
  }
  
  return &S3Client{
    client:   s3.NewFromConfig(awsCfg),
    bucket:   cfg.Bucket,
    provider: cfg.Provider,
  }, nil
}

// GetPresignedUploadURL returns a presigned PUT URL (10 min expiry)
func (sc *S3Client) GetPresignedUploadURL(ctx context.Context, objectKey string) (string, error) {
  presigner := s3.NewPresignClient(sc.client)
  
  presignResult, err := presigner.PresignPutObject(ctx,
    &s3.PutObjectInput{
      Bucket: aws.String(sc.bucket),
      Key:    aws.String(objectKey),
    },
    func(opts *s3.PresignOptions) {
      opts.Expires = time.Duration(10 * time.Minute)
    },
  )
  
  if err != nil {
    return "", fmt.Errorf("presign upload: %w", err)
  }
  
  return presignResult.URL, nil
}

// GetPresignedDownloadURL returns a presigned GET URL (1 hour expiry)
func (sc *S3Client) GetPresignedDownloadURL(ctx context.Context, objectKey string) (string, error) {
  presigner := s3.NewPresignClient(sc.client)
  
  presignResult, err := presigner.PresignGetObject(ctx,
    &s3.GetObjectInput{
      Bucket: aws.String(sc.bucket),
      Key:    aws.String(objectKey),
    },
    func(opts *s3.PresignOptions) {
      opts.Expires = time.Duration(1 * time.Hour)
    },
  )
  
  if err != nil {
    return "", fmt.Errorf("presign download: %w", err)
  }
  
  return presignResult.URL, nil
}

// DeleteObject removes an object from S3
func (sc *S3Client) DeleteObject(ctx context.Context, objectKey string) error {
  _, err := sc.client.DeleteObject(ctx, &s3.DeleteObjectInput{
    Bucket: aws.String(sc.bucket),
    Key:    aws.String(objectKey),
  })
  
  return err
}
```

### 4. gRPC Handler
**File**: `services/documents/internal/infrastructure/grpc/handler.go`

```go
package grpc

import (
  "context"
  "fmt"
  "github.com/google/uuid"
  pb "his/api/his/documents/v1"
  "his/services/documents/internal/domain"
)

type DocumentHandler struct {
  repo domain.DocumentRepository
  s3   *S3Client
  logger *zerolog.Logger
}

func (h *DocumentHandler) InitiateUpload(ctx context.Context, req *pb.InitiateUploadRequest) (*pb.InitiateUploadResponse, error) {
  // 1. Validate auth + RBAC
  ctx, span := tracer.Start(ctx, "DocumentHandler.InitiateUpload")
  defer span.End()
  
  tenantID := uuid.MustParse(req.TenantId)
  // Validate permissions, file size, document type
  
  // 2. Generate S3 object key (tenant-isolated path)
  uploadSessionID := uuid.New()
  objectKey := fmt.Sprintf("%s/documents/%s/%s", tenantID, uploadSessionID, req.FileName)
  
  // 3. Get presigned upload URL
  uploadURL, err := h.s3.GetPresignedUploadURL(ctx, objectKey)
  if err != nil {
    span.RecordError(err)
    return nil, status.Error(codes.Internal, "failed to generate upload URL")
  }
  
  // 4. Audit log
  h.auditLog(ctx, "document.upload.initiated", tenantID, uploadSessionID, req.FileName)
  
  return &pb.InitiateUploadResponse{
    UploadSessionId: uploadSessionID.String(),
    UploadUrl:       uploadURL,
  }, nil
}

func (h *DocumentHandler) CompleteUpload(ctx context.Context, req *pb.CompleteUploadRequest) (*pb.CompleteUploadResponse, error) {
  // 1. Validate auth
  // 2. Lookup document by upload_session_id
  // 3. Create document entity in database
  // 4. Publish domain event (his.documents.document.uploaded)
  // 5. Audit log
  
  return &pb.CompleteUploadResponse{
    DocumentId: docID,
    Metadata:   toProto(doc),
  }, nil
}

func (h *DocumentHandler) GetDownloadUrl(ctx context.Context, req *pb.GetDownloadUrlRequest) (*pb.GetDownloadUrlResponse, error) {
  // 1. Validate auth + fetch document
  // 2. Get presigned download URL
  // 3. Audit log (who downloaded what)
  // 4. Return URL
  
  return &pb.GetDownloadUrlResponse{
    DownloadUrl: downloadURL,
    Metadata:    toProto(doc),
  }, nil
}
```

### 5. Config
**File**: `services/documents/pkg/config/config.go`

```go
package config

type S3Config struct {
  Provider  string // "aws" or "minio"
  Endpoint  string // only for MinIO
  Bucket    string
  Region    string
  AccessKey string // env: S3_ACCESS_KEY
  SecretKey string // env: S3_SECRET_KEY
}

func LoadS3Config() S3Config {
  provider := os.Getenv("S3_PROVIDER")
  if provider == "" {
    provider = "aws"
  }
  
  return S3Config{
    Provider:  provider,
    Endpoint:  os.Getenv("S3_ENDPOINT"),
    Bucket:    os.Getenv("S3_BUCKET"),
    Region:    os.Getenv("AWS_REGION"),
    AccessKey: os.Getenv("S3_ACCESS_KEY"),
    SecretKey: os.Getenv("S3_SECRET_KEY"),
  }
}
```

### 6. Docker Compose (Development)
**File**: `docker-compose.yml`

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: his
      POSTGRES_PASSWORD: his
      POSTGRES_DB: documents

  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data

volumes:
  minio_data:
```

## Environment Variables

```bash
# AWS S3 (Cloud)
S3_PROVIDER=aws
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=***
AWS_SECRET_ACCESS_KEY=***
S3_BUCKET=hospital-documents-prod

# MinIO (On-Premise)
S3_PROVIDER=minio
S3_ENDPOINT=https://minio.hospital.local:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=***
S3_BUCKET=hospital-documents
```

## Testing Strategy

- **Unit tests**: Mock DocumentRepository and S3Client
- **Integration tests**: Use MinIO docker container + testcontainers
- **Presigned URL tests**: Verify URL expiry, signature validation
- **Audit logging**: Verify who/what/when/tenant logged

## Deployment Checklist

**AWS S3**
- [ ] Create S3 bucket (versioning + MFA delete)
- [ ] Configure lifecycle (Glacier after 90 days)
- [ ] Enable access logging
- [ ] Enable encryption (KMS)
- [ ] Block public access
- [ ] Set CloudWatch alarms

**MinIO (On-Prem)**
- [ ] Install MinIO HA (3+ nodes)
- [ ] Enable TLS
- [ ] Configure backup
- [ ] Set retention policies
