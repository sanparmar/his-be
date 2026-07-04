---
name: new-grpc-handler
description: "Generate a gRPC handler method for an existing HIS service with JWT auth, RBAC, OpenTelemetry tracing, audit logging, and error mapping. Use when: adding a new RPC endpoint, implementing a proto service method."
---

# New gRPC Handler

Generate a production-ready gRPC handler method with all required cross-cutting concerns.

## Inputs

- **Service domain**: ${input:domain:Service domain (e.g. patient, billing, appointment)}
- **Entity**: ${input:entity:Entity this handler operates on (e.g. Patient, Invoice)}
- **Operation**: ${input:operation:CRUD operation (Create | Get | Update | Delete | List)}
- **Required role(s)**: ${input:roles:Comma-separated RBAC roles that may call this (e.g. doctor,nurse,admin)}

## Generated Handler

File: `services/${domain}/internal/infrastructure/grpc/server.go` (add method to existing server struct)

### Pattern to Follow

```go
func (s *Server) ${operation}${entity}(
    ctx context.Context,
    req *v1.${operation}${entity}Request,
) (*v1.${operation}${entity}Response, error) {

    // 1. Extract and validate JWT claims
    claims, err := s.auth.ValidateToken(ctx)
    if err != nil {
        return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
    }

    // 2. Enforce RBAC
    if !claims.HasAnyRole(${roles as string slice}) {
        return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
    }

    // 3. Start OTel span
    ctx, span := s.tracer.Start(ctx, "${domain^}Service.${operation}${entity}")
    defer span.End()
    span.SetAttributes(
        attribute.String("tenant.id", claims.TenantID),
        attribute.String("user.id", claims.UserID),
    )

    // 4. Validate request
    if err := s.validator.Struct(req); err != nil {
        span.RecordError(err)
        return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
    }

    // 5. Parse tenant UUID
    tenantID, err := uuid.Parse(claims.TenantID)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
    }

    // 6. Build command/query and dispatch
    cmd := command.${operation}${entity}Command{
        TenantID:  tenantID,
        RequestBy: uuid.MustParse(claims.UserID),
        // map req fields ...
    }
    result, err := s.${operation:lower}${entity}Handler.Handle(ctx, cmd)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        s.logger.Error().Err(err).Str("tenant_id", claims.TenantID).Msg("${operation}${entity} failed")
        return nil, mapDomainError(err)
    }

    // 7. Emit audit log
    s.audit.Log(ctx, audit.Entry{
        TenantID:      tenantID,
        ActorID:       uuid.MustParse(claims.UserID),
        Action:        "${operation:upper}_${entity:upper}",
        ResourceID:    result.ID,
        CorrelationID: extractCorrelationID(ctx),
    })

    // 8. Map to proto response
    return &v1.${operation}${entity}Response{
        ${entity}: mapToProto${entity}(result),
    }, nil
}
```

### Error Mapping Helper
If `mapDomainError` does not exist in the file, generate it:

```go
func mapDomainError(err error) error {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return status.Error(codes.NotFound, "resource not found")
    case errors.Is(err, domain.ErrAlreadyExists):
        return status.Error(codes.AlreadyExists, "resource already exists")
    case errors.Is(err, domain.ErrForbidden):
        return status.Error(codes.PermissionDenied, "operation not permitted")
    default:
        return status.Error(codes.Internal, "internal server error")
    }
}
```

## Unit Test
File: `services/${domain}/internal/infrastructure/grpc/server_test.go` (add test cases)

Generate table-driven test cases for:
- Valid request, happy path
- Unauthenticated (missing/invalid token)
- Insufficient role
- Validation error (missing required field)
- Domain not found error → maps to codes.NotFound
- Internal error → maps to codes.Internal

## Quality Check

- Never return raw error messages that expose internals to the caller
- No PHI in span attributes, log fields, or error messages
- Audit log entry is always written on successful mutation
