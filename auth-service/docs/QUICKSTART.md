# Quick Start Guide

## Prerequisites

- Go 1.23 or higher
- PostgreSQL 15 or higher
- Docker & Docker Compose (optional)

## Method 1: Using Docker Compose (Recommended)

### 1. Start Services

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f auth-service

# Stop services
docker-compose down
```

The service will be available at `http://localhost:8080`

## Method 2: Local Development

### 1. Setup Database

Start PostgreSQL:
```bash
docker-compose up -d auth-db
```

Or use your local PostgreSQL instance.

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your settings:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth_db
JWT_SECRET=your-secret-key-change-this
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Application

```bash
# Using make
make run

# Or directly
go run cmd/server/main.go
```

## Testing the API

### 1. Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "auth-service"
}
```

### 2. Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test@123456"
  }'
```

**Password Requirements**:
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

Expected response:
```json
{
  "message": "user registered successfully"
}
```

### 3. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test@123456"
  }'
```

Expected response:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "abc123...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "...",
    "email": "test@example.com",
    "role": "user",
    "is_verified": false,
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 4. Get Current User (Protected)

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 5. Refresh Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 6. Change Password

```bash
curl -X PUT http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "Test@123456",
    "new_password": "NewTest@123456"
  }'
```

### 7. Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 8. Logout All Devices

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout-all \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## API Endpoints Summary

| Method | Endpoint | Auth Required | Description |
|--------|----------|---------------|-------------|
| GET | `/health` | No | Health check |
| POST | `/api/v1/auth/register` | No | Register new user |
| POST | `/api/v1/auth/login` | No | Login user |
| POST | `/api/v1/auth/refresh` | No | Refresh access token |
| GET | `/api/v1/auth/me` | Yes | Get current user |
| POST | `/api/v1/auth/logout` | Yes | Logout current session |
| POST | `/api/v1/auth/logout-all` | Yes | Logout all sessions |
| PUT | `/api/v1/auth/change-password` | Yes | Change password |

## Development Commands

```bash
# Run application
make run

# Build binary
make build

# Run tests
make test

# Clean build artifacts
make clean

# Download dependencies
make deps

# Docker commands
make docker-up        # Start containers
make docker-down      # Stop containers
make docker-build     # Rebuild images
make docker-logs      # View logs
```

## Database Management

### Connect to PostgreSQL

```bash
# Using docker-compose
docker-compose exec auth-db psql -U postgres -d auth_db

# Or locally
psql -h localhost -U postgres -d auth_db
```

### View Tables

```sql
\dt

-- Output:
--  public | audit_logs      | table | postgres
--  public | refresh_tokens  | table | postgres
--  public | users           | table | postgres
```

### View Users

```sql
SELECT id, email, role, is_verified, is_active, created_at FROM users;
```

### View Audit Logs

```sql
SELECT id, user_id, action, ip_address, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 10;
```

## Troubleshooting

### Connection Refused to Database

Check if PostgreSQL is running:
```bash
docker-compose ps
```

Restart database:
```bash
docker-compose restart auth-db
```

### Port Already in Use

Change port in `.env`:
```env
PORT=8081
```

Or in `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"
```

### JWT Secret Not Set

Make sure `.env` has `JWT_SECRET`:
```env
JWT_SECRET=your-secret-key-at-least-32-characters-long
```

### Migration Issues

Drop and recreate database:
```bash
docker-compose down -v
docker-compose up -d
```

## Testing with Postman

1. Import the following as environment variables:
   - `BASE_URL`: `http://localhost:8080`
   - `ACCESS_TOKEN`: (will be set after login)
   - `REFRESH_TOKEN`: (will be set after login)

2. Create requests:
   - Set `Authorization` header: `Bearer {{ACCESS_TOKEN}}`
   - Set `Content-Type` header: `application/json`

3. Use Tests tab to save tokens:
```javascript
// After login request
pm.environment.set("ACCESS_TOKEN", pm.response.json().access_token);
pm.environment.set("REFRESH_TOKEN", pm.response.json().refresh_token);
```

## Production Deployment Checklist

- [ ] Change `JWT_SECRET` to a strong random value
- [ ] Set `ENVIRONMENT=production`
- [ ] Configure proper `ALLOWED_ORIGINS`
- [ ] Use strong `DB_PASSWORD`
- [ ] Setup HTTPS/TLS
- [ ] Configure proper database backup
- [ ] Setup monitoring and logging
- [ ] Use secrets management (not .env file)
- [ ] Setup rate limiting
- [ ] Configure proper CORS
- [ ] Enable database connection pooling limits
- [ ] Setup health checks for load balancer

## Next Steps

1. Explore the codebase architecture in [ARCHITECTURE.md](ARCHITECTURE.md)
2. Read the full documentation in [README.md](README.md)
3. Add custom features following the Clean Architecture pattern
4. Write tests for your use cases
5. Setup CI/CD pipeline
