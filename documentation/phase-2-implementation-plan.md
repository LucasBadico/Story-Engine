# Phase 2 Implementation Plan: gRPC API Layer

**Goal:** Expose core functionality via gRPC for programmatic access

**Status:** 🔵 Ready to start (Phase 1 complete)

---

## Overview

Phase 2 builds a gRPC API layer on top of the Phase 1 foundation. This will expose story management functionality through a type-safe, high-performance API suitable for service-to-service communication and client applications.

---

## Requirements

### Features
- ✅ gRPC server with graceful shutdown
- ✅ Auth context (tenant + user extraction from metadata)
- ✅ Story management endpoints
- ✅ Tenant management endpoints
- ✅ Error mapping (domain errors → gRPC status codes)

### Proto Definitions
- `Tenant` message
- `User` message
- `Membership` message
- `Story` message
- `Chapter` message
- `Scene` message
- `Beat` message
- `ProseBlock` message
- Common types (timestamps, pagination, errors)

### Endpoints (RPCs)

**TenantService:**
- `CreateTenant` - Create a new tenant
- `GetTenant` - Retrieve tenant by ID

**StoryService:**
- `CreateStory` - Create a new story (version 1)
- `GetStory` - Retrieve story by ID
- `ListStories` - List stories for a tenant (paginated)
- `CloneStory` - Clone a story (create new version)
- `ListStoryVersions` - List all versions of a story

### Technical Scope
- Protobuf v3 definitions
- gRPC server implementation
- Handler layer calling application use cases
- Interceptors for:
  - Logging
  - Error handling
  - Auth context extraction
  - Recovery (panic handling)
- Integration tests with gRPC client

---

## Implementation Steps

### Step 1: Setup gRPC Infrastructure

#### 1.1 Add Dependencies

Update `go.mod` with:
```go
google.golang.org/grpc v1.60.0
google.golang.org/protobuf v1.32.0
```

Install tools:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### 1.2 Create Makefile Targets

Add to `Makefile`:
```makefile
# Proto generation
proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/**/*.proto

proto-clean:
	find proto -name "*.pb.go" -delete
```

---

### Step 2: Define Protobuf Messages

#### 2.1 Common Types

**File:** `proto/common/common.proto`

```protobuf
syntax = "proto3";
package common;
option go_package = "github.com/story-engine/main-service/proto/common";

// Pagination request
message PaginationRequest {
  int32 limit = 1;
  int32 offset = 2;
}

// Error details
message ErrorDetail {
  string code = 1;
  string message = 2;
  map<string, string> metadata = 3;
}
```

#### 2.2 Tenant Messages

**File:** `proto/tenant/tenant.proto`

```protobuf
syntax = "proto3";
package tenant;
option go_package = "github.com/story-engine/main-service/proto/tenant";

import "google/protobuf/timestamp.proto";

message Tenant {
  string id = 1;
  string name = 2;
  string status = 3;
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}

message CreateTenantRequest {
  string name = 1;
  string created_by_user_id = 2; // optional
}

message CreateTenantResponse {
  Tenant tenant = 1;
}

message GetTenantRequest {
  string id = 1;
}

message GetTenantResponse {
  Tenant tenant = 1;
}

service TenantService {
  rpc CreateTenant(CreateTenantRequest) returns (CreateTenantResponse);
  rpc GetTenant(GetTenantRequest) returns (GetTenantResponse);
}
```

#### 2.3 Story Messages

**File:** `proto/story/story.proto`

```protobuf
syntax = "proto3";
package story;
option go_package = "github.com/story-engine/main-service/proto/story";

import "google/protobuf/timestamp.proto";
import "proto/common/common.proto";

message Story {
  string id = 1;
  string tenant_id = 2;
  string title = 3;
  string status = 4;
  int32 version_number = 5;
  string root_story_id = 6;
  string previous_story_id = 7; // optional
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}

message Chapter {
  string id = 1;
  string story_id = 2;
  int32 number = 3;
  string title = 4;
  string status = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}

message Scene {
  string id = 1;
  string story_id = 2;
  string chapter_id = 3;
  int32 number = 4;
  string title = 5;
  string status = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

message CreateStoryRequest {
  string tenant_id = 1;
  string title = 2;
  string created_by_user_id = 3; // optional
}

message CreateStoryResponse {
  Story story = 1;
}

message GetStoryRequest {
  string id = 1;
}

message GetStoryResponse {
  Story story = 1;
}

message ListStoriesRequest {
  string tenant_id = 1;
  common.PaginationRequest pagination = 2;
}

message ListStoriesResponse {
  repeated Story stories = 1;
  int32 total_count = 2;
}

message CloneStoryRequest {
  string source_story_id = 1;
  string created_by_user_id = 2; // optional
}

message CloneStoryResponse {
  Story story = 1;
  int32 new_version_number = 2;
}

message ListStoryVersionsRequest {
  string root_story_id = 1;
}

message ListStoryVersionsResponse {
  repeated Story versions = 1;
}

service StoryService {
  rpc CreateStory(CreateStoryRequest) returns (CreateStoryResponse);
  rpc GetStory(GetStoryRequest) returns (GetStoryResponse);
  rpc ListStories(ListStoriesRequest) returns (ListStoriesResponse);
  rpc CloneStory(CloneStoryRequest) returns (CloneStoryResponse);
  rpc ListStoryVersions(ListStoryVersionsRequest) returns (ListStoryVersionsResponse);
}
```

---

### Step 3: Implement gRPC Server Infrastructure

#### 3.1 Server Setup

**File:** `internal/transport/grpc/server.go`

Core responsibilities:
- Initialize gRPC server with interceptors
- Register services
- Graceful shutdown handling
- Health check endpoint

Key components:
```go
type Server struct {
    grpcServer *grpc.Server
    config     *config.Config
    logger     logger.Logger
}

func NewServer(config *config.Config, logger logger.Logger) *Server
func (s *Server) Start(port int) error
func (s *Server) Stop() error
```

#### 3.2 Interceptors

**File:** `internal/transport/grpc/interceptors/logging.go`
- Log all incoming requests
- Log request duration
- Log errors

**File:** `internal/transport/grpc/interceptors/recovery.go`
- Catch panics
- Convert to gRPC Internal error
- Log stack trace

**File:** `internal/transport/grpc/interceptors/auth.go`
- Extract tenant_id from metadata
- Extract user_id from metadata (optional)
- Inject into context
- Return Unauthenticated if tenant_id missing

**File:** `internal/transport/grpc/interceptors/errors.go`
- Map domain errors to gRPC status codes:
  - `NotFoundError` → `codes.NotFound`
  - `AlreadyExistsError` → `codes.AlreadyExists`
  - `ValidationError` → `codes.InvalidArgument`
  - `UnauthorizedError` → `codes.Unauthenticated`
  - `ForbiddenError` → `codes.PermissionDenied`
  - Default → `codes.Internal`

---

### Step 4: Implement Service Handlers

#### 4.1 Tenant Service Handler

**File:** `internal/transport/grpc/handlers/tenant_handler.go`

```go
type TenantHandler struct {
    pb.UnimplementedTenantServiceServer
    createTenantUseCase *tenant.CreateTenantUseCase
    tenantRepo          repositories.TenantRepository
    logger              logger.Logger
}

func (h *TenantHandler) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest) (*pb.CreateTenantResponse, error)
func (h *TenantHandler) GetTenant(ctx context.Context, req *pb.GetTenantRequest) (*pb.GetTenantResponse, error)
```

Responsibilities:
- Validate protobuf request
- Call application use case
- Map domain model to protobuf message
- Handle errors

#### 4.2 Story Service Handler

**File:** `internal/transport/grpc/handlers/story_handler.go`

```go
type StoryHandler struct {
    pb.UnimplementedStoryServiceServer
    createStoryUseCase  *story.CreateStoryUseCase
    cloneStoryUseCase   *story.CloneStoryUseCase
    versionGraphUseCase *story.GetStoryVersionGraphUseCase
    storyRepo           repositories.StoryRepository
    logger              logger.Logger
}

func (h *StoryHandler) CreateStory(ctx context.Context, req *pb.CreateStoryRequest) (*pb.CreateStoryResponse, error)
func (h *StoryHandler) GetStory(ctx context.Context, req *pb.GetStoryRequest) (*pb.GetStoryResponse, error)
func (h *StoryHandler) ListStories(ctx context.Context, req *pb.ListStoriesRequest) (*pb.ListStoriesResponse, error)
func (h *StoryHandler) CloneStory(ctx context.Context, req *pb.CloneStoryRequest) (*pb.CloneStoryResponse, error)
func (h *StoryHandler) ListStoryVersions(ctx context.Context, req *pb.ListStoryVersionsRequest) (*pb.ListStoryVersionsResponse, error)
```

---

### Step 5: Implement Mappers

**File:** `internal/transport/grpc/mappers/tenant_mapper.go`

```go
func TenantToProto(t *tenant.Tenant) *pb.Tenant
func TenantFromProto(pb *pb.Tenant) *tenant.Tenant
```

**File:** `internal/transport/grpc/mappers/story_mapper.go`

```go
func StoryToProto(s *story.Story) *pb.Story
func ChapterToProto(c *story.Chapter) *pb.Chapter
func SceneToProto(s *story.Scene) *pb.Scene
```

---

### Step 6: Implement Main Entry Point

**File:** `cmd/api-grpc/main.go`

Responsibilities:
- Load configuration
- Setup logger
- Connect to database
- Initialize repositories
- Initialize use cases
- Create handlers
- Start gRPC server
- Handle graceful shutdown (SIGINT, SIGTERM)

---

### Step 7: Context Management

**File:** `internal/transport/grpc/context.go`

```go
type contextKey string

const (
    tenantIDKey contextKey = "tenant_id"
    userIDKey   contextKey = "user_id"
)

func WithTenantID(ctx context.Context, tenantID string) context.Context
func TenantIDFromContext(ctx context.Context) (string, bool)
func WithUserID(ctx context.Context, userID string) context.Context
func UserIDFromContext(ctx context.Context) (string, bool)
```

---

### Step 8: Integration Tests

#### 8.1 Test Infrastructure

**File:** `internal/transport/grpc/testing/client.go`

```go
func SetupTestServer(t *testing.T) (*grpc.ClientConn, func())
func CreateTestClient(conn *grpc.ClientConn) pb.StoryServiceClient
```

#### 8.2 Service Tests

**File:** `internal/transport/grpc/handlers/tenant_handler_test.go`
- Test CreateTenant with valid input
- Test CreateTenant with duplicate name
- Test GetTenant with existing ID
- Test GetTenant with non-existing ID

**File:** `internal/transport/grpc/handlers/story_handler_test.go`
- Test CreateStory with valid input
- Test CreateStory with invalid tenant
- Test GetStory
- Test ListStories with pagination
- Test CloneStory
- Test ListStoryVersions
- Test auth context (tenant_id in metadata)

---

## File Structure

```
main-service/
├── proto/
│   ├── common/
│   │   └── common.proto
│   ├── tenant/
│   │   └── tenant.proto
│   └── story/
│       └── story.proto
├── internal/
│   └── transport/
│       └── grpc/
│           ├── server.go
│           ├── context.go
│           ├── interceptors/
│           │   ├── logging.go
│           │   ├── recovery.go
│           │   ├── auth.go
│           │   └── errors.go
│           ├── handlers/
│           │   ├── tenant_handler.go
│           │   ├── tenant_handler_test.go
│           │   ├── story_handler.go
│           │   └── story_handler_test.go
│           ├── mappers/
│           │   ├── tenant_mapper.go
│           │   └── story_mapper.go
│           └── testing/
│               └── client.go
└── cmd/
    └── api-grpc/
        └── main.go
```

---

## Implementation Checklist

### Infrastructure
- [ ] Add gRPC dependencies to go.mod
- [ ] Create proto generation Makefile targets
- [ ] Setup common proto types
- [ ] Create server infrastructure

### Proto Definitions
- [ ] Define common.proto (pagination, errors)
- [ ] Define tenant.proto (Tenant, TenantService)
- [ ] Define story.proto (Story, Chapter, Scene, StoryService)
- [ ] Generate Go code from protos

### Interceptors
- [ ] Implement logging interceptor
- [ ] Implement recovery interceptor
- [ ] Implement auth context interceptor
- [ ] Implement error mapping interceptor

### Handlers
- [ ] Implement TenantHandler (CreateTenant, GetTenant)
- [ ] Implement StoryHandler (CreateStory, GetStory, ListStories, CloneStory, ListStoryVersions)
- [ ] Create domain → proto mappers
- [ ] Create proto → domain mappers (if needed)

### Main Application
- [ ] Implement cmd/api-grpc/main.go
- [ ] Graceful shutdown handling
- [ ] Configuration loading
- [ ] Dependency injection

### Testing
- [ ] Create test infrastructure (test server, client)
- [ ] Test TenantHandler endpoints
- [ ] Test StoryHandler endpoints
- [ ] Test auth context injection
- [ ] Test error mapping
- [ ] Test interceptors

### Documentation
- [ ] Update README with gRPC setup instructions
- [ ] Document API endpoints
- [ ] Document metadata requirements (tenant_id, user_id)
- [ ] Create example client code

---

## Success Criteria

### Functional Requirements
- ✅ gRPC server starts and accepts connections
- ✅ All 7 RPC methods implemented and working
- ✅ Auth context (tenant_id) extracted from metadata
- ✅ Domain errors properly mapped to gRPC status codes
- ✅ Validation errors return InvalidArgument
- ✅ Not found errors return NotFound
- ✅ Multi-tenancy isolation enforced

### Non-Functional Requirements
- ✅ Integration tests pass with real gRPC client
- ✅ Graceful shutdown works correctly
- ✅ Panics are recovered and logged
- ✅ All requests/responses logged
- ✅ No linter errors
- ✅ Code follows Clean Architecture principles

### Documentation
- ✅ Proto files well-documented
- ✅ README updated with gRPC instructions
- ✅ Example client code provided

---

## Testing Strategy

### Unit Tests
- Handler logic (mocked use cases)
- Mapper functions
- Context utilities

### Integration Tests
- Full gRPC server with real database
- Test all endpoints end-to-end
- Test error scenarios
- Test auth context propagation

### Manual Testing
Use `grpcurl` for manual testing:

```bash
# CreateTenant
grpcurl -plaintext \
  -d '{"name": "Test Tenant"}' \
  localhost:9090 \
  tenant.TenantService/CreateTenant

# CreateStory (with tenant_id in metadata)
grpcurl -plaintext \
  -H "tenant_id: <tenant-uuid>" \
  -d '{"title": "My Story"}' \
  localhost:9090 \
  story.StoryService/CreateStory
```

---

## Dependencies to Add

```bash
go get google.golang.org/grpc@v1.60.0
go get google.golang.org/protobuf@v1.32.0
go get google.golang.org/genproto/googleapis/rpc/errdetails@latest
```

---

## Configuration

Add to `internal/platform/config/config.go`:

```go
type GRPCConfig struct {
    Port            int    `env:"GRPC_PORT" default:"9090"`
    MaxRecvMsgSize  int    `env:"GRPC_MAX_RECV_MSG_SIZE" default:"4194304"`  // 4MB
    MaxSendMsgSize  int    `env:"GRPC_MAX_SEND_MSG_SIZE" default:"4194304"`  // 4MB
    EnableReflection bool  `env:"GRPC_ENABLE_REFLECTION" default:"true"`
}
```

---

## Notes

### Phase 2 Scope Decisions

**In Scope:**
- TenantService (CreateTenant, GetTenant)
- StoryService (CreateStory, GetStory, ListStories, CloneStory, ListStoryVersions)
- Auth context via metadata (tenant_id, user_id)
- Basic interceptors (logging, recovery, auth, errors)

**Out of Scope (Future Phases):**
- User authentication (JWT validation)
- Authorization (role-based access control)
- Chapter/Scene/Beat management endpoints
- Streaming RPCs
- Rate limiting
- API versioning

### Design Decisions

1. **Auth Context via Metadata:** Simple approach for Phase 2. tenant_id is required in metadata for story operations. user_id is optional.

2. **Error Mapping:** Domain errors are mapped to appropriate gRPC status codes via interceptor, keeping handlers clean.

3. **Pagination:** Simple offset-based pagination for Phase 2. Can be enhanced with cursor-based pagination later.

4. **Validation:** Input validation happens in handlers before calling use cases. Proto validation (protoc-gen-validate) can be added later.

5. **Testing:** Integration tests use a real gRPC server with test database, ensuring end-to-end correctness.

---

## Next Steps After Phase 2

Once Phase 2 is complete, the system will have:
- ✅ Complete backend (domain + persistence + API)
- ✅ gRPC API for programmatic access
- ✅ Foundation for client applications

**Phase 3** will focus on building an Obsidian plugin as the first client application.

---

## Estimated Effort

- **Proto Definitions:** 2-3 hours
- **Server Infrastructure:** 3-4 hours
- **Interceptors:** 2-3 hours
- **Handlers:** 4-5 hours
- **Mappers:** 1-2 hours
- **Main Entry Point:** 1-2 hours
- **Integration Tests:** 3-4 hours
- **Documentation:** 1-2 hours

**Total:** ~18-25 hours

