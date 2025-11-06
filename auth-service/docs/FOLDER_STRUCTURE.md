# Folder Structure Explained

## Overview

```
auth-service/
├── cmd/                    # Application entry points
├── internal/               # Private application code (Clean Architecture layers)
├── pkg/                    # Public reusable packages (CÓ THỂ dùng ở services khác)
├── scripts/                # Build, deployment, utility scripts
├── docs/                   # Documentation files
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── README.md
└── go.mod
```

## Chi tiết từng folder

### 1. `cmd/` - Application Entry Points

**Mục đích**: Chứa các main packages để build executables.

```
cmd/
└── server/
    └── main.go         # HTTP server entry point
```

**Đặc điểm**:
- Mỗi subfolder là một executable
- Chỉ chứa dependency injection và wiring
- Code tối thiểu, chỉ khởi tạo và chạy app

**Ví dụ thêm commands**:
```
cmd/
├── server/             # HTTP server
├── worker/             # Background worker (nếu cần)
└── migrate/            # Database migration CLI (nếu tách riêng)
```

---

### 2. `internal/` - Private Application Code

**Mục đích**: Code riêng của service này, KHÔNG được import bởi services khác.

```
internal/
├── domain/             # Layer 1: Business logic core
├── application/        # Layer 2: Use cases
├── infrastructure/     # Layer 3: Technical implementations
└── delivery/          # Layer 4: API/UI layer
```

**Tại sao internal?**
- Go compiler tự động enforce: code trong `internal/` không thể import từ bên ngoài module
- Đảm bảo encapsulation giữa các services
- Tránh coupling giữa microservices

#### 2.1. `internal/domain/` - Domain Layer

**Pure business logic**, không phụ thuộc framework hay infrastructure.

```
domain/
├── entity/             # Domain entities với business rules
├── repository/         # Repository interfaces (data access contracts)
├── service/           # Domain service interfaces
└── errors/            # Domain-specific errors
```

**Đặc điểm**:
- ✅ Business rules
- ✅ Interfaces definition
- ❌ KHÔNG có database code
- ❌ KHÔNG có HTTP code
- ❌ KHÔNG import infrastructure

#### 2.2. `internal/application/` - Application Layer

**Orchestrate business logic**, implement use cases.

```
application/
├── dto/               # Data Transfer Objects (request/response)
└── usecase/          # Use cases (business workflows)
```

**Đặc điểm**:
- ✅ Use domain entities và interfaces
- ✅ Coordinate giữa repositories và services
- ❌ KHÔNG biết về HTTP hay database specifics

#### 2.3. `internal/infrastructure/` - Infrastructure Layer

**Technical implementations**, external dependencies.

```
infrastructure/
├── config/            # Configuration management
├── logger/            # Logging implementation
├── persistence/       # Database implementations
│   └── postgres/
└── security/         # Security implementations (JWT, Password)
```

**TẠI SAO config và logger ở đây thay vì `pkg/`?**

**Lý do 1: Infrastructure Concerns**
- Config và logger là **infrastructure concerns**, không phải business logic
- Chúng implement cách service tương tác với external systems
- Trong Clean Architecture, đây là outer layer

**Lý do 2: Service-Specific Implementation**
```go
// infrastructure/config/config.go
type Config struct {
    Environment string
    Server      ServerConfig
    Database    DatabaseConfig      // ← Specific cho service này
    JWT         JWTConfig          // ← Specific cho auth service
    Security    SecurityConfig     // ← Specific cho auth service
}
```

Config này chứa **business-specific settings** (JWT, Security), không phải generic config.

**Lý do 3: Dependency Direction**
```
internal/infrastructure/ implements internal/domain/ interfaces
```

Nếu để ở `pkg/`, sẽ bị ngược chiều dependency.

**Khi nào NÊN để ở `pkg/`?**
```go
// pkg/config/loader.go - Generic config loader
func LoadFromEnv(prefix string) (map[string]string, error) {
    // Generic environment variable loader
}

// pkg/logger/interface.go - Generic logger interface
type Logger interface {
    Info(msg string, fields ...Field)
    Error(msg string, fields ...Field)
}
```

#### 2.4. `internal/delivery/` - Delivery Layer

**Handle external communication** (HTTP, gRPC, CLI, etc.)

```
delivery/
└── http/
    ├── handler/       # HTTP request handlers
    ├── middleware/    # HTTP middleware
    └── router/       # Route definitions
```

**Đặc điểm**:
- ✅ Framework-specific code (Gin)
- ✅ Request/response transformation
- ✅ Call use cases
- ❌ KHÔNG có business logic

---

### 3. `pkg/` - Public Reusable Packages

**Mục đích**: Code CÓ THỂ được reuse bởi **services khác** trong cùng organization.

```
pkg/
├── response/          # HTTP response helpers
├── validator/         # Generic validators
└── utils/            # Utility functions
```

**Nguyên tắc cho pkg/**:
1. **Generic** - Không specific cho một service
2. **Reusable** - Có thể dùng ở nhiều services
3. **Stable** - API ít thay đổi
4. **No business logic** - Chỉ là utilities

**Ví dụ ĐÚNG cho pkg/:**
```go
// pkg/response/response.go
package response

// Generic HTTP response wrapper
func JSON(c *gin.Context, status int, data interface{}) {
    c.JSON(status, Response{Data: data})
}
```

**Ví dụ SAI (nên ở internal/):**
```go
// pkg/auth/jwt.go ❌ SAI
package auth

// Too specific, should be in internal/infrastructure/security/
type JWTService struct {
    secret string
}
```

**So sánh pkg/ vs internal/:**

| Aspect | `pkg/` | `internal/` |
|--------|--------|-------------|
| Visibility | Public, có thể import từ services khác | Private, chỉ service này dùng |
| Purpose | Generic utilities | Business logic + implementation |
| Stability | Phải stable, nhiều service depend | Có thể thay đổi thoải mái |
| Examples | HTTP helpers, validators, utils | Use cases, repositories, entities |

**Khi nào tạo pkg/?**
- Khi bạn copy-paste code giống nhau giữa nhiều services
- Khi bạn muốn share utilities với other teams
- Khi code là generic và không chứa business logic

---

### 4. `scripts/` - Automation Scripts

**Mục đích**: Scripts cho development, testing, deployment.

```
scripts/
├── migrate.sh         # Run database migrations
├── seed.sh           # Seed test data
└── test.sh           # Run tests with coverage
```

**Ví dụ thêm scripts:**
```
scripts/
├── build.sh          # Build for different platforms
├── deploy.sh         # Deployment script
├── backup.sh         # Database backup
└── docker-build.sh   # Docker build script
```

---

### 5. `docs/` - Documentation

**Mục đích**: Tất cả documentation files.

```
docs/
├── ARCHITECTURE.md              # Architecture explanation
├── API.md                       # API documentation
├── QUICKSTART.md               # Quick start guide
├── IMPLEMENTATION_SUMMARY.md   # Implementation details
└── FOLDER_STRUCTURE.md         # This file
```

**Tại sao tách docs/ folder?**
- ✅ Root folder gọn gàng hơn
- ✅ Dễ tìm documentation
- ✅ Có thể generate docs site (MkDocs, Docusaurus)
- ✅ Tách biệt docs với code

---

## Best Practices

### 1. Dependency Rule

```
cmd → internal/delivery → internal/application → internal/domain
                ↓              ↓
         internal/infrastructure
```

**Rule**: Inner layers KHÔNG phụ thuộc outer layers.

### 2. Import Rules

**✅ ĐƯỢC PHÉP:**
```go
// internal/delivery/http/handler/auth_handler.go
import (
    "auth-service/internal/application/usecase"  // ✅ OK
    "auth-service/pkg/response"                  // ✅ OK
)
```

**❌ KHÔNG ĐƯỢC:**
```go
// internal/domain/entity/user.go
import (
    "auth-service/internal/infrastructure/logger"  // ❌ WRONG!
    "github.com/gin-gonic/gin"                    // ❌ WRONG!
)
```

### 3. Khi nào tạo pkg/ mới?

**Hỏi 3 câu:**
1. Code này có generic không? (không specific cho auth service)
2. Code này có thể reuse ở service khác không?
3. Code này có chứa business logic không? (không nên có)

Nếu 1=YES, 2=YES, 3=NO → Đặt vào `pkg/`
Ngược lại → Đặt vào `internal/`

---

## Examples

### Example 1: Thêm Email Service

**Nếu generic email sender:**
```
pkg/
└── email/
    ├── sender.go      # Generic SMTP sender
    └── template.go    # Generic template engine
```

**Nếu specific cho auth (verification emails):**
```
internal/
└── infrastructure/
    └── notification/
        └── email_service.go  # Auth-specific email logic
```

### Example 2: Thêm Metrics

**Generic metrics collector:**
```
pkg/
└── metrics/
    ├── prometheus.go
    └── collector.go
```

**Auth-specific metrics:**
```
internal/
└── infrastructure/
    └── metrics/
        └── auth_metrics.go  # Track login attempts, etc.
```

---

## Migration from Old Structure

**Old structure:**
```
auth-service-old/
├── config/          # ← Mixed
├── internal/
│   ├── model/      # ← Mixed with entity
│   ├── service/    # ← Mixed with usecase
│   └── repository/ # ← OK
└── utils/          # ← Should be pkg/
```

**New structure (Clean Architecture):**
```
auth-service/
├── internal/
│   ├── domain/
│   │   ├── entity/         # ← Pure entities
│   │   └── repository/     # ← Interfaces only
│   ├── application/
│   │   └── usecase/        # ← Business workflows
│   └── infrastructure/
│       ├── config/         # ← Service-specific config
│       └── persistence/    # ← Repository implementations
└── pkg/
    └── utils/              # ← Generic utilities
```

---

## Summary

### `internal/` vs `pkg/` Decision Tree

```
Is this code specific to auth-service?
├─ YES → internal/
│   └─ Does it contain business logic?
│       ├─ YES → internal/domain/ or internal/application/
│       └─ NO → internal/infrastructure/
│
└─ NO → Can it be reused by other services?
    ├─ YES → pkg/
    └─ NO → internal/
```

### Key Points

1. **`internal/infrastructure/`** - Service-specific technical implementations
   - Config with business settings (JWT, Security)
   - Logger với service context
   - Database models và repositories

2. **`pkg/`** - Generic, reusable utilities
   - HTTP response helpers
   - Generic validators
   - Common utilities

3. **`scripts/`** - Automation và tooling
   - Database migrations
   - Testing scripts
   - Deployment automation

4. **`docs/`** - Documentation
   - Giữ root folder clean
   - Easy to maintain
   - Can generate doc sites

**Nguyên tắc vàng**: 
- Bắt đầu với `internal/`, chỉ move sang `pkg/` khi thực sự cần share
- Prefer duplication over wrong abstraction
- Keep business logic in `internal/domain/`
