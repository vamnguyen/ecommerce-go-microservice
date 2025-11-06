# Auth Service

A clean architecture authentication microservice built with Go, featuring JWT-based authentication, refresh tokens, and comprehensive security features.

## Architecture

This service follows Clean Architecture principles with clear separation of concerns:

```
auth-service/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── domain/          # Business logic & entities
│   │   ├── entity/      # Domain entities
│   │   ├── repository/  # Repository interfaces
│   │   ├── service/     # Service interfaces
│   │   └── errors/      # Domain errors
│   ├── application/     # Application business rules
│   │   ├── dto/         # Data transfer objects
│   │   └── usecase/     # Use cases/interactors
│   ├── infrastructure/  # External interfaces
│   │   ├── config/      # Configuration
│   │   ├── logger/      # Logging
│   │   ├── security/    # JWT & password services
│   │   └── persistence/ # Database implementations
│   └── delivery/        # Delivery mechanisms
│       └── http/        # HTTP handlers & routes
│           ├── handler/
│           ├── middleware/
│           └── router/
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Features

- **User Registration & Authentication**
  - Email/password registration
  - Login with credential validation
  - Account lockout after failed attempts
  - Password strength validation

- **Token Management**
  - JWT access tokens (short-lived)
  - Refresh tokens (long-lived)
  - Token rotation on refresh
  - Revoke tokens (logout)
  - Revoke all user tokens (logout all)

- **Security**
  - Bcrypt password hashing
  - Account lockout mechanism
  - Audit logging
  - CORS support
  - Request validation

- **Clean Architecture**
  - Domain-driven design
  - Dependency injection
  - Interface-based design
  - Testable & maintainable

## Prerequisites

- Go 1.23+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

## Quick Start

### Using Docker Compose

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

### Local Development

1. **Setup Database**
   ```bash
   docker-compose up -d postgres
   ```

2. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Run Application**
   ```bash
   make deps
   make run
   ```

## API Endpoints

### Public Endpoints

- `GET /health` - Health check
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token

### Protected Endpoints (Require Authentication)

- `GET /api/v1/auth/me` - Get current user
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/logout-all` - Logout from all devices
- `PUT /api/v1/auth/change-password` - Change password

## API Examples

### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "StrongP@ss123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "StrongP@ss123"
  }'
```

### Get Current User
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

## Configuration

Key environment variables:

```env
# Server
PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth_db

# JWT
JWT_SECRET=your-secret-key
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=720h

# Security
MAX_LOGIN_ATTEMPTS=5
ACCOUNT_LOCK_DURATION=15m
```

## Development

```bash
# Install dependencies
make deps

# Run application
make run

# Build binary
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## Docker

```bash
# Build image
make docker-build

# Start services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## Database Schema

The service uses PostgreSQL with the following tables:

- **users** - User accounts
- **refresh_tokens** - Refresh token records
- **audit_logs** - Security audit trail

## Security Features

1. **Password Security**
   - Minimum 8 characters
   - Must contain uppercase, lowercase, number, and special character
   - Bcrypt hashing with cost factor 10

2. **Account Protection**
   - Failed login tracking
   - Automatic account lockout
   - Configurable lockout duration

3. **Token Security**
   - Short-lived access tokens
   - Refresh token rotation
   - Token revocation support

4. **Audit Logging**
   - All authentication events logged
   - IP address and user agent tracking
   - Metadata support for additional context

## License

MIT
