---
applyTo: "proto/**/*.proto"
---

# Proto File Conventions

## Syntax and Package
Always use proto3. Package and Go package must follow:

```proto
syntax = "proto3";

package his.<domain>.v1;

option go_package = "github.com/his-platform/his-be/api/v1/<domain>";
```

## Service Naming
- Service: `<Domain>Service`
- Methods: `Create<Entity>`, `Get<Entity>`, `Update<Entity>`, `Delete<Entity>`, `List<Entity>s`
- Request/Response: `<Method>Request` / `<Method>Response`

## Standard Request Pattern
```proto
message Create<Entity>Request {
  string tenant_id  = 1; // UUID as string
  string created_by = 2; // UUID as string
  // domain fields ...
}

message Create<Entity>Response {
  <Entity> <entity> = 1;
}
```

## Standard Entity Message Pattern
```proto
message <Entity> {
  string id         = 1;
  string tenant_id  = 2;
  // domain fields (3+) ...
  string created_at = 98; // RFC3339
  string updated_at = 99; // RFC3339
}
```

## Pagination for List Methods
```proto
message List<Entity>sRequest {
  string tenant_id  = 1;
  int32  page_size  = 2;
  string page_token = 3;
  // filter fields ...
}

message List<Entity>sResponse {
  repeated <Entity> items          = 1;
  string            next_page_token = 2;
  int64             total_count    = 3;
}
```

## PHI Fields in Proto
- PHI field names must be clearly marked with a comment: `// PHI - encrypted at rest`
- Never expose PHI in error messages or status details

## Field Numbering
- Reserve fields 1–10 for core identifiers (id, tenant_id, created_by, etc.)
- Reserve fields 90–99 for audit timestamps
- Never reuse or remove field numbers once published

## Health Check
Every service proto MUST include the standard health check:
```proto
import "google/protobuf/empty.proto";

rpc HealthCheck(google.protobuf.Empty) returns (HealthCheckResponse);

message HealthCheckResponse {
  string status  = 1; // "ok" | "degraded"
  string version = 2;
}
```

## Versioning
- All proto files live under `proto/v1/`
- Breaking changes require a new `v2/` directory — never modify published v1 messages
