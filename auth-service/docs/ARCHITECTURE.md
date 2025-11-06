# Auth Service - Clean Architecture

## Overview

Auth Service là một microservice xác thực người dùng được xây dựng theo nguyên tắc Clean Architecture, đảm bảo tính module, dễ test và dễ bảo trì.

## Architecture Layers

### 1. Domain Layer (Lớp Nghiệp vụ)
**Mục đích**: Chứa business logic thuần túy, không phụ thuộc vào framework hay công nghệ bên ngoài.

```
internal/domain/
├── entity/              # Domain entities
│   ├── user.go         # User entity với business rules
│   ├── refresh_token.go
│   └── audit_log.go
├── repository/         # Repository interfaces
│   ├── user_repository.go
│   ├── refresh_token_repository.go
│   └── audit_log_repository.go
├── service/           # Domain service interfaces
│   ├── password_service.go
│   └── token_service.go
└── errors/           # Domain errors
    └── errors.go
```

**Đặc điểm**:
- Entities chứa business logic (IsAccountLocked, UpdateLastLogin, etc.)
- Interfaces định nghĩa contracts cho repositories và services
- Không có dependencies đến layers khác
- Pure Go code, không import framework

### 2. Application Layer (Lớp Ứng dụng)
**Mục đích**: Orchestrate business logic, implement use cases.

```
internal/application/
├── dto/               # Data Transfer Objects
│   └── auth_dto.go   # Request/Response models
└── usecase/          # Use cases/Interactors
    └── auth_usecase.go
```

**Đặc điểm**:
- Use cases implement business workflows
- Sử dụng domain entities và repository interfaces
- Transform giữa DTOs và domain entities
- Không phụ thuộc vào infrastructure hay delivery

### 3. Infrastructure Layer (Lớp Hạ tầng)
**Mục đích**: Implement technical details, external dependencies.

```
internal/infrastructure/
├── config/           # Configuration management
│   └── config.go
├── logger/           # Logging implementation
│   └── logger.go
├── persistence/      # Database implementations
│   └── postgres/
│       ├── database.go
│       ├── user_repository.go
│       ├── refresh_token_repository.go
│       └── audit_log_repository.go
└── security/        # Security implementations
    ├── jwt_service.go
    └── password_service.go
```

**Đặc điểm**:
- Implement domain interfaces (repositories, services)
- Chứa database models và ORM logic
- Implement JWT, password hashing, logging
- Có thể thay đổi implementation mà không ảnh hưởng domain

### 4. Delivery Layer (Lớp Giao tiếp)
**Mục đích**: Handle external communication (HTTP, gRPC, etc.).

```
internal/delivery/
└── http/
    ├── handler/      # HTTP handlers
    │   ├── auth_handler.go
    │   └── health_handler.go
    ├── middleware/   # HTTP middleware
    │   ├── auth_middleware.go
    │   ├── cors_middleware.go
    │   ├── logger_middleware.go
    │   └── recovery_middleware.go
    └── router/      # Route definitions
        └── router.go
```

**Đặc điểm**:
- HTTP handlers gọi use cases
- Middleware xử lý cross-cutting concerns
- Request validation và response formatting
- Framework-specific code (Gin)

## Dependency Flow

```
┌─────────────────────────────────────────────┐
│           Delivery Layer (HTTP)             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Handler  │  │Middleware│  │  Router  │ │
│  └──────────┘  └──────────┘  └──────────┘ │
└───────────────────┬─────────────────────────┘
                    │ depends on
                    ▼
┌─────────────────────────────────────────────┐
│         Application Layer (Use Cases)       │
│  ┌──────────┐            ┌──────────┐      │
│  │ Use Case │            │   DTOs   │      │
│  └──────────┘            └──────────┘      │
└───────────────────┬─────────────────────────┘
                    │ depends on
                    ▼
┌─────────────────────────────────────────────┐
│            Domain Layer (Core)              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Entity   │  │Repository│  │  Service │ │
│  │          │  │Interface │  │Interface │ │
│  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────────────────────┘
                    ▲
                    │ implements
                    │
┌─────────────────────────────────────────────┐
│      Infrastructure Layer (Technical)       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │Repository│  │  Service │  │  Config  │ │
│  │  Impl    │  │   Impl   │  │  Logger  │ │
│  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────────────────────┘
```

## Key Design Principles

### 1. Dependency Inversion
- Domain layer không phụ thuộc vào bất kỳ layer nào
- Outer layers phụ thuộc vào inner layers
- Infrastructure implements domain interfaces

### 2. Single Responsibility
- Mỗi component có một trách nhiệm rõ ràng
- Entities chứa business rules
- Use cases orchestrate workflows
- Repositories handle data access

### 3. Interface Segregation
- Interfaces nhỏ, tập trung (UserRepository, TokenService)
- Dễ mock cho testing
- Dễ thay đổi implementation

### 4. Separation of Concerns
- Business logic tách biệt khỏi technical details
- HTTP handlers chỉ lo việc HTTP
- Database logic chỉ ở repository implementations

## Data Flow Example: User Login

```
1. HTTP Request
   ↓
2. AuthHandler.Login()
   ↓
3. Validate request (DTO binding)
   ↓
4. AuthUseCase.Login()
   ├─→ UserRepository.FindByEmail() (check user exists)
   ├─→ PasswordService.VerifyPassword() (verify password)
   ├─→ User.UpdateLastLogin() (domain logic)
   ├─→ TokenService.GenerateAccessToken()
   ├─→ TokenService.GenerateRefreshToken()
   ├─→ RefreshTokenRepository.Create()
   └─→ AuditLogRepository.Create()
   ↓
5. Return AuthResponse (DTO)
   ↓
6. HTTP Response (JSON)
```

## Testing Strategy

### Unit Testing
- **Domain Entities**: Test business logic methods
- **Use Cases**: Mock repositories and services
- **Handlers**: Mock use cases

### Integration Testing
- **Repository Implementations**: Test with real database
- **API Endpoints**: Test full request/response cycle

### Example:
```go
// Mock repository for testing use case
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    args := m.Called(ctx, email)
    return args.Get(0).(*entity.User), args.Error(1)
}

// Test use case
func TestAuthUseCase_Login(t *testing.T) {
    mockUserRepo := new(MockUserRepository)
    mockPasswordService := new(MockPasswordService)
    // ... setup mocks and test
}
```

## Scalability Features

### 1. Horizontal Scaling
- Stateless service (JWT tokens)
- Database connection pooling
- Mỗi instance độc lập

### 2. Microservice Ready
- Self-contained service
- Clear API boundaries
- Can be deployed independently
- Database per service pattern

### 3. Performance
- Database indexes trên user.email, refresh_tokens.token_hash
- Connection pooling configuration
- Efficient bcrypt cost factor

### 4. Monitoring
- Structured logging (zap)
- Audit trail cho security events
- Health check endpoints

## Configuration Management

```go
// Environment-based configuration
type Config struct {
    Environment string
    Server      ServerConfig
    Database    DatabaseConfig
    JWT         JWTConfig
    Security    SecurityConfig
}
```

**Best Practices**:
- Sensitive data qua environment variables
- Validation khi load config
- Default values hợp lý
- Production-ready defaults

## Security Features

### 1. Password Security
- Bcrypt hashing (cost 10)
- Password strength validation
- No password in logs/responses

### 2. Token Security
- Short-lived access tokens (15m)
- Long-lived refresh tokens (30d)
- Token rotation on refresh
- Token revocation support

### 3. Account Protection
- Failed login tracking
- Account lockout (5 attempts, 15m)
- Audit logging

### 4. API Security
- CORS configuration
- Input validation
- Error message sanitization

## Adding New Features

### Example: Add Email Verification

**1. Domain Layer**:
```go
// entity/user.go
type User struct {
    // ... existing fields
    VerificationToken *string
    VerifiedAt        *time.Time
}

func (u *User) Verify() {
    now := time.Now()
    u.VerifiedAt = &now
    u.VerificationToken = nil
}
```

**2. Repository Interface**:
```go
// domain/repository/user_repository.go
type UserRepository interface {
    // ... existing methods
    FindByVerificationToken(ctx context.Context, token string) (*User, error)
}
```

**3. Use Case**:
```go
// application/usecase/verify_email_usecase.go
func (uc *AuthUseCase) VerifyEmail(ctx context.Context, token string) error {
    user, err := uc.userRepo.FindByVerificationToken(ctx, token)
    if err != nil {
        return err
    }
    user.Verify()
    return uc.userRepo.Update(ctx, user)
}
```

**4. Handler**:
```go
// delivery/http/handler/auth_handler.go
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
    token := c.Query("token")
    err := h.authUseCase.VerifyEmail(c.Request.Context(), token)
    // ... handle response
}
```

**5. Route**:
```go
// delivery/http/router/router.go
auth.GET("/verify-email", r.authHandler.VerifyEmail)
```

## Maintenance Benefits

### 1. Testability
- Each layer can be tested independently
- Easy to mock dependencies
- Fast unit tests without database

### 2. Flexibility
- Easy to swap implementations (Postgres → MySQL)
- Add new delivery methods (gRPC, GraphQL)
- Change frameworks without touching business logic

### 3. Readability
- Clear structure, easy to navigate
- Code organized by concern
- Self-documenting architecture

### 4. Maintainability
- Changes isolated to specific layers
- Reduced coupling
- Easier onboarding for new developers
