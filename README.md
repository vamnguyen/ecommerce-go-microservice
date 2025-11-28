# E-commerce Go Microservice MVP

Dự án e-commerce backend sử dụng kiến trúc microservice với Go, gRPC, và Kong API Gateway.

## Kiến trúc

Dự án sử dụng **Clean Architecture** với các tầng:

- **Domain Layer**: Entities, Repository interfaces, Domain services
- **Application Layer**: Use cases, DTOs
- **Infrastructure Layer**: Database, gRPC handlers, External clients
- **Delivery Layer**: gRPC handlers, Interceptors

## Services

### 1. Auth Service (Port 9002)

- Quản lý authentication và authorization
- JWT với RS256
- Refresh token với token family
- Account locking sau nhiều lần đăng nhập sai
- Audit logging

**Endpoints:**

- `POST /api/v1/auth/register` - Đăng ký user mới
- `POST /api/v1/auth/login` - Đăng nhập
- `POST /api/v1/auth/refresh` - Refresh token
- `GET /api/v1/auth/me` - Lấy thông tin user hiện tại
- `POST /api/v1/auth/logout` - Đăng xuất
- `POST /api/v1/auth/logout-all` - Đăng xuất tất cả thiết bị
- `POST /api/v1/auth/change-password` - Đổi mật khẩu

### 2. User Service (Port 9003)

- Quản lý user profiles
- CRUD operations cho user profile

**Endpoints:**

- `GET /api/v1/users/profile` - Lấy profile của user hiện tại
- `PUT /api/v1/users/profile` - Cập nhật profile
- `GET /api/v1/users/{user_id}` - Lấy thông tin user
- `GET /api/v1/users` - List users (với pagination)

### 3. Order Service (Port 9004)

- Quản lý orders
- Tích hợp với user-service để validate users
- Order status management

**Endpoints:**

- `POST /api/v1/orders` - Tạo order mới
- `GET /api/v1/orders/{order_id}` - Lấy thông tin order
- `GET /api/v1/orders` - List orders của user
- `PATCH /api/v1/orders/{order_id}/status` - Cập nhật status

## Công nghệ sử dụng

- **Go 1.24**: Ngôn ngữ lập trình
- **gRPC**: Communication protocol giữa các services
- **Kong Gateway**: API Gateway với JWT authentication
- **PostgreSQL**: Database cho mỗi service
- **GORM**: ORM cho Go
- **OpenTelemetry**: Distributed tracing và metrics
- **Prometheus**: Metrics collection
- **Jaeger**: Tracing visualization
- **Grafana**: Metrics visualization
- **Docker & Docker Compose**: Containerization

## Cấu trúc thư mục

```
ecommerce-go-microservice/
├── auth-service/          # Authentication service
├── user-service/          # User profile service
├── order-service/         # Order management service
├── api-gateway/           # Kong configuration
├── observability/         # Prometheus config
├── proto-common/          # Shared proto files (google/api)
└── docker-compose.yml     # Docker orchestration
```

Mỗi service có cấu trúc:

```
service-name/
├── cmd/server/            # Entry point
├── internal/
│   ├── domain/           # Domain entities, repositories
│   ├── application/      # Use cases, DTOs
│   ├── delivery/grpc/    # gRPC handlers
│   └── infrastructure/   # Database, clients, config
├── proto/                # gRPC proto definitions
└── gen/go/              # Generated gRPC code
```

## Setup và chạy

### Prerequisites

- Docker & Docker Compose
- Go 1.24+ (nếu build local)
- Make (optional)

### Chạy với Docker Compose

```bash
# Start tất cả services
docker-compose up -d

# Xem logs
docker-compose logs -f

# Stop tất cả services
docker-compose down
```

### Build và chạy local

Mỗi service có Makefile riêng:

```bash
# Generate proto files
cd auth-service && make proto
cd user-service && make proto
cd order-service && make proto

# Build
make build

# Run
make run
```

## API Gateway

Kong Gateway chạy trên port **8000**:

- API Gateway: `http://localhost:8000`
- Kong Admin: `http://localhost:8001`

Tất cả requests phải đi qua Kong Gateway. JWT token được validate bởi Kong trước khi forward đến services.

## Observability

- **Jaeger UI**: `http://localhost:16686`
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000` (admin/admin)

## Database

Mỗi service có database riêng:

- `auth-db`: Port 5432
- `user-db`: Port 5433
- `order-db`: Port 5434

## Kiến thức học được

Dự án này giúp học:

1. **Clean Architecture**: Tách biệt các tầng, dependency inversion
2. **Microservices**: Service decomposition, inter-service communication
3. **gRPC**: Protocol buffers, streaming, error handling
4. **API Gateway**: Request routing, authentication, rate limiting
5. **Distributed Systems**: Service discovery, health checks, circuit breakers
6. **Observability**: Tracing, metrics, logging
7. **Security**: JWT, token refresh, account locking
8. **Database Design**: Per-service database, migrations
9. **Docker**: Containerization, orchestration
10. **Go Best Practices**: Error handling, context, graceful shutdown

## Setup và Build

### 1. Cài đặt Dependencies

Có 2 cách để cài đặt dependencies:

**Cách 1: Cài đặt cho tất cả services (khuyến nghị)**

```bash
# Từ root directory
make deps
# hoặc
./scripts/install-deps.sh
```

**Cách 2: Cài đặt cho từng service**

```bash
# Auth service
cd auth-service && make deps

# User service
cd user-service && make deps

# Order service
cd order-service && make deps
```

**Lệnh Go trực tiếp:**

```bash
cd auth-service && go mod download && go mod tidy
cd user-service && go mod download && go mod tidy
cd order-service && go mod download && go mod tidy
```

### 2. Generate Proto Files

Trước khi build, cần generate proto files cho tất cả services:

```bash
# Tất cả services
make proto

# Hoặc từng service
cd auth-service && make proto
cd user-service && make proto
cd order-service && make proto
```

**Lưu ý**: Tất cả services sử dụng `proto-common/google/api` để import `google/api/annotations.proto`. Thư mục `proto-common/` được mount vào Kong container tại `/etc/kong/proto/google`.

### JWT Keys

Auth service cần JWT keys trong `auth-service/certs/`:

- `private_key.pem`
- `public_key.pem`

### Database Migrations

Database migrations tự động chạy khi service khởi động (GORM AutoMigrate).

## Next Steps

Để mở rộng MVP:

1. Thêm Product Service
2. Thêm Payment Service
3. Thêm Inventory Service
4. Implement event-driven architecture với message queue
5. Thêm caching layer (Redis)
6. Implement rate limiting
7. Thêm comprehensive testing
8. CI/CD pipeline
9. Service mesh (Istio)
10. API versioning
