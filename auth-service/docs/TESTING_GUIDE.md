# Testing Guide - Kong JWT Authentication

## Overview
Auth-service giờ chỉ chạy gRPC server (port 9002). Kong API Gateway sẽ:
- Validate JWT token cho protected routes
- Inject user_id vào gRPC metadata
- Forward HTTP requests sang gRPC calls

## Setup

### 1. Start Services
```bash
# Terminal 1: Start Kong
cd /path/to/ecommerce-go-microservice
docker-compose up kong auth-db

# Terminal 2: Start auth-service (gRPC only)
cd auth-service
go run cmd/server/main.go
```

### 2. Verify Services
```bash
# Check Kong
curl http://localhost:8000

# Check gRPC (nếu có grpcurl)
grpcurl -plaintext localhost:9002 list
```

## Testing Flow

### Step 1: Register User (Public Route)
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

Expected response:
```json
{
  "message": "user registered successfully"
}
```

### Step 2: Login (Public Route)
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

Expected response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "base64encodedrefreshtoken..."
}
```

**⚠️ Save the access_token for next steps!**

### Step 3: Get User Info (Protected Route)
```bash
ACCESS_TOKEN="your_access_token_from_login"

curl http://localhost:8000/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Expected response:
```json
{
  "id": "uuid-here",
  "email": "test@example.com",
  "role": "user",
  "is_verified": false,
  "is_active": true,
  "created_at": "2024-11-10T10:00:00Z"
}
```

### Step 4: Refresh Token (Protected Route)
```bash
REFRESH_TOKEN="your_refresh_token_from_login"

curl -X POST http://localhost:8000/api/v1/auth/refresh \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'$REFRESH_TOKEN'",
    "access_token": "'$ACCESS_TOKEN'"
  }'
```

Expected response:
```json
{
  "access_token": "new_access_token...",
  "refresh_token": "new_refresh_token..."
}
```

### Step 5: Change Password (Protected Route)
```bash
curl -X POST http://localhost:8000/api/v1/auth/change-password \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "SecurePass123!",
    "new_password": "NewSecurePass456!"
  }'
```

Expected response:
```json
{
  "message": "password changed successfully"
}
```

### Step 6: Logout (Protected Route)
```bash
curl -X POST http://localhost:8000/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'$REFRESH_TOKEN'"
  }'
```

Expected response: `204 No Content`

### Step 7: Logout All Devices (Protected Route)
```bash
curl -X POST http://localhost:8000/api/v1/auth/logout-all \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Expected response: `204 No Content`

## Error Cases

### 1. Missing Authorization Header (Protected Routes)
```bash
curl http://localhost:8000/api/v1/auth/me
```
Expected: `401 Unauthorized` - JWT validation failed

### 2. Invalid Token
```bash
curl http://localhost:8000/api/v1/auth/me \
  -H "Authorization: Bearer invalid_token"
```
Expected: `401 Unauthorized` - JWT validation failed

### 3. Expired Token
Use an old/expired token:
```bash
curl http://localhost:8000/api/v1/auth/me \
  -H "Authorization: Bearer expired_token"
```
Expected: `401 Unauthorized` - Token expired

### 4. Access Protected Route Without Token
```bash
curl http://localhost:8000/api/v1/auth/me
```
Expected: `401 Unauthorized`

## Debugging

### Check Kong JWT Configuration
```bash
curl http://localhost:8001/routes
curl http://localhost:8001/plugins
```

### Check gRPC Server Logs
Auth-service logs sẽ show:
- Request metadata (x-user-id, x-user-email)
- gRPC method calls
- Errors

### Verify JWT Secret
Đảm bảo JWT_SECRET trong:
- `/auth-service/.env`: `your-secret-key-change-this-in-production`
- `/api-gateway/kong.yml` consumer secret: `your-secret-key-change-this-in-production`

Phải **GIỐNG NHAU**!

## Architecture Flow

```
Client (HTTP)
    ↓
Kong Gateway (Port 8000)
    ↓
JWT Plugin validates token
    ↓
Request Transformer adds x-user-id header
    ↓
gRPC-Gateway Plugin
    ↓
Auth-Service gRPC (Port 9002)
    ↓
Auth Interceptor extracts x-user-id from metadata
    ↓
Business Logic
```

## Notes

- **Public Routes**: `/register`, `/login`, `/health` - No JWT required
- **Protected Routes**: All others - JWT required in `Authorization: Bearer <token>` header
- **Token Expiry**: Access token = 15 minutes, Refresh token = 30 days (configurable)
- **Kong validates JWT locally** - No call to auth-service for validation
- **Auth-service only used for**: Register, Login, RefreshToken, Business logic
