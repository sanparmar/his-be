---
name: write-unit-tests
description: "Generate comprehensive table-driven Go unit tests with mocks for a HIS service file. Use when: adding tests to a new or existing Go file, improving test coverage, testing domain logic or handlers."
---

# Write Unit Tests

Generate complete table-driven Go unit tests for the target file.

## Inputs

- **File to test**: ${input:targetFile:Relative path to the Go file (e.g. services/patient/internal/domain/service.go)}
- **Coverage target**: ${input:coverage:Minimum coverage % to aim for (default: 80)}

## Instructions

1. **Read the target file** thoroughly before generating any tests.
2. Identify all exported functions, methods, and types.
3. For each function/method, generate a `TestXxx` function with table-driven cases.

## Test Structure Template

```go
package <same_package>_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

func Test<FunctionName>(t *testing.T) {
    tenantID := uuid.New()
    actorID  := uuid.New()

    tests := []struct {
        name      string
        setup     func(mockRepo *MockRepository)
        input     <InputType>
        wantErr   bool
        wantErrIs error
        validate  func(t *testing.T, result <OutputType>)
    }{
        {
            name: "success - <happy path description>",
            setup: func(mockRepo *MockRepository) {
                mockRepo.On("<Method>", mock.Anything, mock.AnythingOfType("uuid.UUID")).
                    Return(<valid return>, nil)
            },
            input:   <valid input>,
            wantErr: false,
            validate: func(t *testing.T, result <OutputType>) {
                assert.Equal(t, tenantID, result.TenantID)
                // assert domain invariants
            },
        },
        {
            name: "error - not found",
            setup: func(mockRepo *MockRepository) {
                mockRepo.On("<Method>", mock.Anything, mock.Anything).
                    Return(nil, domain.ErrNotFound)
            },
            input:     <input triggering not found>,
            wantErr:   true,
            wantErrIs: domain.ErrNotFound,
        },
        {
            name: "error - validation: missing tenant_id",
            setup: func(mockRepo *MockRepository) {},
            input:   <input with zero TenantID>,
            wantErr: true,
        },
        // ... edge cases
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            tc.setup(mockRepo)

            svc := New<Service>(mockRepo, zerolog.Nop(), nopTracer())
            result, err := svc.<FunctionName>(context.Background(), tc.input)

            if tc.wantErr {
                require.Error(t, err)
                if tc.wantErrIs != nil {
                    assert.ErrorIs(t, err, tc.wantErrIs)
                }
                return
            }
            require.NoError(t, err)
            if tc.validate != nil {
                tc.validate(t, result)
            }
            mockRepo.AssertExpectations(t)
        })
    }
}
```

## Coverage Requirements

Ensure test cases cover:
- [ ] Happy path with valid inputs
- [ ] Each error branch (repository errors, domain errors)
- [ ] Boundary conditions (empty strings, zero UUIDs, nil pointers)
- [ ] Multi-tenant isolation (wrong tenant should not access another tenant's data)
- [ ] Soft delete behavior if applicable
- [ ] PHI field handling — assert PHI never appears in error messages or logs

## Mock Generation

If mocks do not exist yet, include a comment at the top of the test file:
```go
//go:generate mockery --name=<RepositoryInterface> --output=./mocks --outpkg=mocks
```

Use `testify/mock` pattern for all mocks. Never mock concrete structs — only interfaces.

## Helper Utilities

Generate these helpers if not already present in a `testutils_test.go` in the same package:

```go
func nopTracer() trace.Tracer {
    return otel.Tracer("nop")
}

func newTestTenantID() uuid.UUID { return uuid.New() }
func newTestActorID()  uuid.UUID { return uuid.New() }
```
