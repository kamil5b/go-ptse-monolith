# Unit Testing Guide

This document describes the unit testing approach, patterns, and best practices used throughout this project.

## Table of Contents

1. [Overview](#overview)
2. [Test Structure](#test-structure)
3. [Table-Driven Tests](#table-driven-tests)
4. [Mocking Strategy](#mocking-strategy)
5. [Common Patterns](#common-patterns)
6. [Running Tests](#running-tests)
7. [Test Coverage](#test-coverage)
8. [Examples](#examples)

## Overview

This project uses Go's standard testing package (`testing`) combined with:

- **GoMock** (`github.com/golang/mock/gomock`) - For generating and managing mocks
- **Testify** - For assertions (`assert`, `require`)
- **Table-driven tests** - For organizing multiple test cases per function

### Testing Philosophy

- **100% coverage goal** - Critical paths should have high test coverage
- **Per-method organization** - One main test function per method with table-driven cases
- **Idiomatic Go** - Following Go testing conventions and best practices
- **Maintainability** - Tests should be easy to read, extend, and debug

## Test Structure

### File Naming Convention

Test files follow Go's standard naming convention:
- Source file: `service.go`
- Test file: `service_test.go`

Example structure:
```
internal/modules/product/service/v1/
├── service_v1.product.go          # Implementation
├── service_v1_test.go             # Tests
└── mocks/
    └── mock_interfaces.go         # Generated mocks
```

### Test File Organization

Each test file is organized as follows:

```go
package v1

import (
    "context"
    "testing"
    
    "github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "go-modular-monolith/internal/modules/product/domain"
    "go-modular-monolith/internal/modules/product/domain/mocks"
)

// Context key type (avoid collisions with built-in types)
type contextKey struct{}

var txContextKey = contextKey{}

// Table-driven tests for each exported method
func TestServiceV1_MethodName(t *testing.T) { ... }

// Benchmark tests
func BenchmarkServiceV1_MethodName(b *testing.B) { ... }
```

## Table-Driven Tests

### What Are Table-Driven Tests?

Table-driven tests organize multiple test cases for a single function into a data structure (usually a slice of structs). This approach:

- Reduces code duplication
- Makes it easy to add new test cases
- Improves readability
- Follows Go idioms

### Basic Pattern

```go
func TestServiceV1_Create(t *testing.T) {
    tests := []struct {
        name      string
        // Input parameters
        req       *domain.CreateProductRequest
        createdBy string
        // Expected outcomes
        wantErr bool
        wantID  string
    }{
        {
            name: "success scenario",
            req: &domain.CreateProductRequest{
                Name:        "Product",
                Description: "Description",
            },
            createdBy: "user123",
            wantErr:   false,
            wantID:    "prod123",
        },
        {
            name:    "error scenario",
            req:     nil,
            wantErr: true,
            wantID:  "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            // Execution
            result, err := service.Create(ctx, tt.req, tt.createdBy)
            
            // Assertions
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.wantID, result.ID)
            }
        })
    }
}
```

### Naming Test Cases

Use descriptive names that clearly state what's being tested:

```go
tests := []struct {
    name string
    // ...
}{
    {name: "success"},
    {name: "success with event bus"},
    {name: "repository error"},
    {name: "product not found"},
    {name: "partial update"},
    {name: "validation error"},
}
```

## Mocking Strategy

### Mock Generation

Mocks are generated using GoMock. See [MOCKS.md](./MOCKS.md) for detailed setup instructions.

Generated mocks provide:
- Full interface implementation
- Call expectation setting
- Argument matching
- Call recording

### Using Mocks in Tests

#### 1. Create Controller

```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()  // Validates all expectations were met
```

#### 2. Create Mocks

```go
mockRepo := mocks.NewMockRepository(ctrl)
mockEventBus := mocks.NewMockEventBus(ctrl)
```

#### 3. Set Expectations

```go
// Method will be called exactly once
mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(1)

// Method can be called any number of times
mockEventBus.EXPECT().Publish(ctx, gomock.Any()).Return(nil).AnyTimes()

// Method will never be called
mockRepo.EXPECT().Delete(ctx, "id").Times(0)
```

#### 4. Custom Behavior with DoAndReturn

```go
mockRepo.EXPECT().GetByID(ctx, "prod123").DoAndReturn(func(c context.Context, id string) (*Product, error) {
    assert.Equal(t, "prod123", id)
    return &Product{ID: "prod123", Name: "Test"}, nil
}).Times(1)
```

### Argument Matching

GoMock provides flexible argument matching:

```go
// Match any value
gomock.Any()

// Match specific value
"exact-string"
123

// Match with custom function
gomock.Not("value")
gomock.Nil()
gomock.Not(gomock.Nil())
```

### When to Use Mocks

- **Always mock external dependencies** (database, APIs, message queues)
- **Isolate unit under test** - Test only the service/handler logic
- **Don't mock what you're testing** - Only mock dependencies
- **Use nil for optional dependencies** - If EventBus is optional, pass nil instead of creating unnecessary mocks

## Common Patterns

### Pattern 1: Success Path

```go
{
    name: "success",
    input: "valid",
    wantErr: false,
    want: expectedResult,
},
```

Setup:
```go
mockRepo.EXPECT().GetByID(ctx, "id").Return(product, nil).Times(1)
// ... execute
assert.Equal(t, expectedResult, actual)
```

### Pattern 2: Error Handling

```go
{
    name: "database error",
    input: "valid",
    repoErr: errors.New("connection failed"),
    wantErr: true,
},
```

Setup:
```go
mockRepo.EXPECT().GetByID(ctx, "id").Return(nil, errors.New("connection failed")).Times(1)
// ... execute
assert.Error(t, err)
assert.Equal(t, "connection failed", err.Error())
```

### Pattern 3: Not Found / Missing Data

```go
{
    name: "product not found",
    id: "nonexistent",
    wantErr: true,
    wantErrMsg: "not found",
},
```

### Pattern 4: Partial Updates

```go
{
    name: "partial update (only name)",
    update: &UpdateRequest{Name: "New Name"},
    wantName: "New Name",
    wantDesc: originalDesc,  // Should remain unchanged
},
```

### Pattern 5: Event Publishing

```go
mockEventBus.EXPECT().Publish(txCtx, gomock.MatcherFunc(func(e interface{}) bool {
    event, ok := e.(domain.ProductCreatedEvent)
    return ok && event.ProductID == "prod123"
})).Times(1)
```

## Running Tests

### Run All Tests

```bash
# Verbose output
go test ./... -v

# With coverage
go test ./... -v -cover

# Specific coverage threshold
go test ./... -cover -coverprofile=coverage.out
```

### Run Tests for Specific Package

```bash
# Service tests only
go test ./internal/modules/product/service/v1/ -v

# All product module tests
go test ./internal/modules/product/... -v
```

### Run Specific Test Function

```bash
# Single test
go test ./internal/modules/product/service/v1/ -run TestServiceV1_Create -v

# Subtest
go test ./internal/modules/product/service/v1/ -run TestServiceV1_Create/success -v
```

### Benchmark Tests

```bash
# Run benchmarks
go test ./internal/modules/product/service/v1/ -bench=. -benchmem

# Run specific benchmark
go test ./internal/modules/product/service/v1/ -bench=BenchmarkServiceV1_Create -benchmem
```

### Run with Timeout

```bash
# 10 second timeout (useful for detecting hanging tests)
go test ./... -timeout 10s -v
```

## Test Coverage

### Checking Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
```

### Coverage Goals

This project aims for:
- **Core business logic**: 90%+ coverage
- **Service/handler layers**: 80%+ coverage
- **Repository layer**: 70%+ coverage (often limited by database drivers)
- **Overall**: Maintain 80%+ project coverage

### What to Test

✅ **DO Test:**
- Happy path scenarios
- Error conditions
- Edge cases (empty lists, nil values, boundary values)
- State transitions
- Business logic
- Event publishing

❌ **DON'T Test:**
- Third-party library implementations
- Generated code (mocks)
- Trivial getters/setters
- Framework-specific routing (very thoroughly)
- External API calls (use mocks instead)

## Examples

### Example 1: Service Create Method

**Source Code** (`service_v1.product.go`):
```go
func (s *ServiceV1) Create(ctx context.Context, req *domain.CreateProductRequest, createdBy string) (*domain.Product, error) {
    ctx = s.uow.StartContext(ctx)
    var p domain.Product
    p.Name = req.Name
    p.Description = req.Description
    p.CreatedAt = time.Now().UTC()
    p.CreatedBy = createdBy
    
    err := s.repo.Create(ctx, &p)
    if err != nil {
        s.uow.DeferErrorContext(ctx, err)
        return nil, err
    }

    if s.eventBus != nil {
        _ = s.eventBus.Publish(ctx, domain.ProductCreatedEvent{...})
    }

    s.uow.DeferErrorContext(ctx, nil)
    return &p, nil
}
```

**Test Code** (`service_v1_test.go`):
```go
func TestServiceV1_Create(t *testing.T) {
    tests := []struct {
        name    string
        req     *domain.CreateProductRequest
        wantErr bool
        wantID  string
    }{
        {
            name: "success",
            req: &domain.CreateProductRequest{
                Name:        "Test Product",
                Description: "A test product",
            },
            wantErr: false,
            wantID:  "prod123",
        },
        {
            name:    "repository error",
            req:     &domain.CreateProductRequest{Name: "Product"},
            repoErr: errors.New("database error"),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockRepo := mocks.NewMockRepository(ctrl)
            mockUOW := mocks.NewMockUnitOfWork(ctrl)
            mockEventBus := mocks.NewMockEventBus(ctrl)

            ctx := context.Background()
            txCtx := context.WithValue(ctx, txKey, "transaction")

            mockUOW.EXPECT().StartContext(ctx).Return(txCtx).Times(1)

            if tt.repoErr != nil {
                mockRepo.EXPECT().Create(txCtx, gomock.Any()).Return(tt.repoErr).Times(1)
                mockUOW.EXPECT().DeferErrorContext(txCtx, tt.repoErr).Return(nil).Times(1)
            } else {
                mockRepo.EXPECT().Create(txCtx, gomock.Any()).DoAndReturn(func(c context.Context, p *domain.Product) error {
                    p.ID = tt.wantID
                    return nil
                }).Times(1)
                mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Times(1)
                mockUOW.EXPECT().DeferErrorContext(txCtx, nil).Return(nil).Times(1)
            }

            service := NewServiceV1(mockRepo, mockUOW, mockEventBus)
            product, err := service.Create(ctx, tt.req, "user123")

            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, product)
            } else {
                require.NoError(t, err)
                require.NotNil(t, product)
                assert.Equal(t, tt.wantID, product.ID)
            }
        })
    }
}
```

### Example 2: Handler GET Request

**Test Pattern for HTTP Handler:**
```go
func TestHandler_GetProduct(t *testing.T) {
    tests := []struct {
        name           string
        productID      string
        mockProduct    *domain.Product
        mockErr        error
        wantStatusCode int
        wantBody       string
    }{
        {
            name:      "success",
            productID: "prod123",
            mockProduct: &domain.Product{ID: "prod123", Name: "Product"},
            wantStatusCode: http.StatusOK,
        },
        {
            name:           "not found",
            productID:      "nonexistent",
            mockErr:        errors.New("not found"),
            wantStatusCode: http.StatusNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockService := mocks.NewMockService(ctrl)
            mockService.EXPECT().Get(gomock.Any(), tt.productID).Return(tt.mockProduct, tt.mockErr).Times(1)

            handler := NewHandler(mockService)
            req := httptest.NewRequest("GET", "/products/"+tt.productID, nil)
            w := httptest.NewRecorder()

            handler.GetProduct(w, req)

            assert.Equal(t, tt.wantStatusCode, w.Code)
        })
    }
}
```

## Best Practices

### 1. Use `require` for Setup, `assert` for Assertions

```go
// Setup - use require (fails test immediately if setup fails)
require.NoError(t, err)
require.NotNil(t, result)

// Assertions - use assert (logs error but continues)
assert.Equal(t, expected, actual)
assert.True(t, condition)
```

### 2. Keep Mocks Focused

```go
// ✅ GOOD - Clear what's being tested
mockRepo.EXPECT().Create(ctx, product).Return(nil).Times(1)

// ❌ BAD - Too general, hard to debug
mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
```

### 3. Test Behavior, Not Implementation

```go
// ✅ GOOD - Tests behavior (what the method does)
assert.Equal(t, createdAt, result.CreatedAt)

// ❌ BAD - Tests implementation details
assert.True(t, time.Now().Before(result.CreatedAt))
```

### 4. Use Descriptive Test Names

```go
// ✅ GOOD - Clear intent
{name: "returns error when product not found"},
{name: "publishes event after successful creation"},

// ❌ BAD - Vague
{name: "error case"},
{name: "test 1"},
```

### 5. Organize Test Data

```go
// ✅ GOOD - Clear separation of test cases
tests := []struct {
    name     string
    input    *Request
    wantErr  bool
    wantCode int
}{
    // success cases first
    // then error cases
}

// ❌ BAD - Mixed, hard to follow
tests := []struct{
    // cases jumbled together
}
```

## HTTP Handler Testing

### Overview

HTTP handlers are typically tested using table-driven tests with mocked dependencies (service layer). The handler's responsibility is to:

1. Extract/bind request data
2. Call the service layer
3. Return appropriate HTTP responses

### Handler Test Pattern

The standard pattern for testing HTTP handlers follows this structure:

```go
func TestHandler_MethodName(t *testing.T) {
    cases := []struct {
        name  string
        setup func(svc *mockdomain.MockService, mc *ctxmocks.MockContext)
    }{
        {
            name: "scenario name",
            setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
                // Set expectations on mocks
            },
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            svc := mockdomain.NewMockService(ctrl)
            mc := ctxmocks.NewMockContext(ctrl)
            h := NewHandler(svc)

            tc.setup(svc, mc)

            if err := h.MethodName(mc); err != nil {
                t.Fatalf("MethodName returned error: %v", err)
            }
        })
    }
}
```

### Handler Test Components

#### 1. Mock Context Object

The mock context represents the HTTP context (Echo, Gin, Fiber, etc.). Common methods:

```go
// Data binding
mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.RequestType{})).Return(nil)

// Getting context
mc.EXPECT().GetContext().Return(context.Background())

// Getting request parameters/body
mc.EXPECT().Param("id").Return("value")
mc.EXPECT().Get("key").Return(value)

// Getting authenticated user
mc.EXPECT().GetUserID().Return("user123")

// Sending responses
mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
```

#### 2. Setup Function

The `setup` function in each test case configures mock expectations. This keeps test cases clean and organized:

```go
setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
    // Set expectations in order they'll be called
    mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.Request{})).Return(nil)
    mc.EXPECT().GetContext().Return(context.Background())
    mc.EXPECT().GetUserID().Return("user1")
    svc.EXPECT().Create(gomock.Any(), gomock.Any(), "user1").Return(&domain.Product{ID: "p1"}, nil)
    mc.EXPECT().JSON(http.StatusCreated, gomock.Any()).Return(nil)
},
```

### Common HTTP Handler Scenarios

#### Scenario 1: Successful Request with Data Binding

**Use Case:** POST/PUT requests that require request body validation

```go
{
    name: "ok",
    setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
        mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(nil)
        mc.EXPECT().GetContext().Return(context.Background())
        mc.EXPECT().GetUserID().Return("user1")
        svc.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.CreateProductRequest{}), "user1").
            Return(&domain.Product{ID: "p1"}, nil)
        mc.EXPECT().JSON(http.StatusCreated, gomock.Any()).Return(nil)
    },
},
```

**What's tested:**
- Request binding succeeds
- Service is called with correct arguments
- Response status and data are sent

#### Scenario 2: Request Binding Error

**Use Case:** Malformed request body or invalid input

```go
{
    name: "bind error",
    setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
        mc.EXPECT().GetContext().Return(context.Background())
        mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(errors.New("bad input"))
        mc.EXPECT().JSON(http.StatusBadRequest, gomock.Any()).Return(nil)
    },
},
```

**What's tested:**
- Handler catches binding errors
- Returns 400 Bad Request status
- Service is NOT called

#### Scenario 3: Service Returns Error

**Use Case:** Business logic error (validation failure, database error, etc.)

```go
{
    name: "service error",
    setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
        mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(nil)
        mc.EXPECT().GetContext().Return(context.Background())
        mc.EXPECT().GetUserID().Return("user1")
        svc.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.CreateProductRequest{}), "user1").
            Return(nil, errors.New("boom"))
        mc.EXPECT().JSON(http.StatusInternalServerError, gomock.Any()).Return(nil)
    },
},
```

**What's tested:**
- Service error is caught and handled
- Returns 500 Internal Server Error status
- Error doesn't crash the handler

#### Scenario 4: GET Request with Path Parameter

**Use Case:** Retrieving a resource by ID

```go
{
    name: "ok",
    setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
        mc.EXPECT().GetContext().Return(context.Background())
        mc.EXPECT().Param("id").Return("p1")
        svc.EXPECT().Get(gomock.Any(), "p1").Return(&domain.Product{ID: "p1"}, nil)
        mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
    },
},
```

**What's tested:**
- Path parameters are extracted correctly
- Service is called with correct ID
- Response is returned with success status

#### Scenario 5: Resource Not Found

**Use Case:** GET/DELETE on non-existent resource

```go
{
    name: "not found",
    setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
        mc.EXPECT().GetContext().Return(context.Background())
        mc.EXPECT().Param("id").Return("p1")
        svc.EXPECT().Get(gomock.Any(), "p1").Return(nil, errors.New("not found"))
        mc.EXPECT().JSON(http.StatusNotFound, gomock.Any()).Return(nil)
    },
},
```

**What's tested:**
- 404 Not Found is returned for missing resources
- Error message is appropriate

#### Scenario 6: Optional Context Values

**Use Case:** User ID may or may not be present in context

```go
{
    name: "ok_with_user",
    setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
        mc.EXPECT().GetContext().Return(context.Background())
        mc.EXPECT().Param("id").Return("p1")
        mc.EXPECT().Get("user_id").Return("user1")  // Returns user ID
        svc.EXPECT().Delete(gomock.Any(), "p1", "user1").Return(nil)
        mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
    },
},
{
    name: "ok_without_user",
    setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
        mc.EXPECT().GetContext().Return(context.Background())
        mc.EXPECT().Param("id").Return("p1")
        mc.EXPECT().Get("user_id").Return(nil)  // No user ID
        svc.EXPECT().Delete(gomock.Any(), "p1", "").Return(nil)
        mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
    },
},
```

**What's tested:**
- Handler works with and without optional context values
- Service receives empty string when value is not present

### Best Practices for Handler Tests

1. **Set expectations in order** - Mock expectations should match the order they'll be called
2. **Use `Bind` early** - If binding fails, service shouldn't be called
3. **Use `AssignableToTypeOf`** - For matching request struct types without knowing exact values
4. **Test error paths** - Always test binding errors and service errors
5. **Keep setup functions focused** - Each case should test one specific scenario
6. **Don't test framework behavior** - Focus on your handler logic, not Echo/Gin internals

### Real-World Example: Product Handler

```go
func TestHandler_Create(t *testing.T) {
    cases := []struct {
        name  string
        setup func(svc *mockdomain.MockService, mc *ctxmocks.MockContext)
    }{
        // Success case: all validations pass
        {
            name: "ok",
            setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
                mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(nil)
                mc.EXPECT().GetContext().Return(context.Background())
                mc.EXPECT().GetUserID().Return("user1")
                svc.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.CreateProductRequest{}), "user1").
                    Return(&domain.Product{ID: "p1"}, nil)
                mc.EXPECT().JSON(http.StatusCreated, gomock.Any()).Return(nil)
            },
        },
        // Error case: invalid request body
        {
            name: "bind error",
            setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
                mc.EXPECT().GetContext().Return(context.Background())
                mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).
                    Return(errors.New("bad input"))
                mc.EXPECT().JSON(http.StatusBadRequest, gomock.Any()).Return(nil)
            },
        },
        // Error case: service layer fails
        {
            name: "service error",
            setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
                mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(nil)
                mc.EXPECT().GetContext().Return(context.Background())
                mc.EXPECT().GetUserID().Return("user1")
                svc.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.CreateProductRequest{}), "user1").
                    Return(nil, errors.New("boom"))
                mc.EXPECT().JSON(http.StatusInternalServerError, gomock.Any()).Return(nil)
            },
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            svc := mockdomain.NewMockService(ctrl)
            mc := ctxmocks.NewMockContext(ctrl)
            h := NewHandler(svc)

            tc.setup(svc, mc)

            if err := h.Create(mc); err != nil {
                t.Fatalf("Create returned error: %v", err)
            }
        })
    }
}
```

## Troubleshooting

### Tests Fail with "missing call(s)"

This means a mock expectation was set but the method wasn't called. Check:
1. Is the method actually called in the code?
2. Is the method called with the expected arguments?
3. Is the mock passed to the service correctly?

### Tests Hang or Timeout

This usually means:
1. A goroutine is waiting for a channel that's never written to
2. A mock expectation is never satisfied
3. Use `go test -timeout 10s` to catch hanging tests

### Nil Pointer Dereference

Common causes:
1. Not initializing mock dependencies properly
2. Passing nil to a method that shouldn't receive nil
3. Not checking for nil in the implementation

### "Unfulfilled expectations"

The test ended but not all mock expectations were satisfied. Verify:
1. Are you calling `ctrl.Finish()` in defer?
2. Are your `.Times()` counts correct?
3. Have you removed a test case but left its expectations?

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [GoMock Documentation](https://github.com/golang/mock)
- [Testify Assertions](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)
