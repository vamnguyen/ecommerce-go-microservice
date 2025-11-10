# Implementation Summary - Kong JWT Authentication

## 🎯 Objective
Implement API Gateway JWT validation pattern để:
- Kong Gateway validate JWT thay vì auth-service
- Auth-service chỉ chạy gRPC (không còn HTTP)
- Giảm load cho auth-service (không phải validate mọi request)
- Microservices nhận user_id từ metadata (không cần validate JWT)

## ✅ Changes Made

### 1. Auth Service - Proto Updates
**File**: `auth-service/proto/auth.proto`
- ✅ Added gRPC methods:
  - `RefreshToken` - Refresh access token
  - `Logout` - Logout single device
  - `LogoutAll` - Logout all devices
  - `GetMe` - Get current user info
  - `ChangePassword` - Change user password
- ✅ Removed `ip_address` and `user_agent` from `LoginRequest` (Kong will inject via headers)
- ✅ Added request/response messages for all new methods

### 2. Auth Service - gRPC Interceptor
**File**: `auth-service/internal/delivery/grpc/interceptor/auth_interceptor.go` (NEW)
- ✅ Extract `x-user-id`, `x-user-email` from gRPC metadata
- ✅ Extract `x-forwarded-for`, `user-agent` for audit logs
- ✅ Public methods bypass authentication (Register, Login, HealthCheck)
- ✅ Protected methods require `x-user-id` in metadata
- ✅ Helper functions: `GetUserIDFromContext()`, `GetClientIPFromContext()`, etc.

### 3. Auth Service - gRPC Handler
**File**: `auth-service/internal/delivery/grpc/handler/grpc_handler.go`
- ✅ Implemented all new gRPC methods
- ✅ Use interceptor helpers to get user_id from context
- ✅ Removed JWT validation logic (Kong handles it)
- ✅ Updated error handling to use existing `toGRPCError()` function

**File**: `auth-service/internal/delivery/grpc/handler/error_handler.go`
- ✅ Added `ErrMissingToken` to error mapping

### 4. Auth Service - Main Server
**File**: `auth-service/cmd/server/main.go`
- ✅ Removed HTTP server completely
- ✅ Only gRPC server on port 9002
- ✅ Added auth interceptor to gRPC server
- ✅ Simplified graceful shutdown (single server)
- ✅ Removed imports: `net/http`, `sync`, HTTP delivery packages

### 5. Auth Service - Cleanup
- ✅ Deleted entire `internal/delivery/http` folder (handlers, middleware, router)
- ✅ Regenerated proto files with new methods

### 6. Kong Gateway Configuration
**File**: `api-gateway/kong.yml`
- ✅ Added JWT consumer with shared secret
  ```yaml
  consumers:
    - username: system
      jwt_secrets:
        - key: system
          secret: your-secret-key-change-this-in-production
          algorithm: HS256
  ```

- ✅ Split routes into public and protected:
  - **Public routes**: `/register`, `/login`, `/health` - No JWT
  - **Protected routes**: `/refresh`, `/logout`, `/logout-all`, `/me`, `/change-password` - Require JWT

- ✅ JWT Plugin on protected routes:
  ```yaml
  - name: jwt
    config:
      secret_is_base64: false
      claims_to_verify: [exp]
      key_claim_name: user_id
  ```

- ✅ Request Transformer to inject headers:
  ```yaml
  - name: request-transformer
    config:
      add:
        headers:
          - x-user-id:$(claims.user_id)
          - x-user-email:$(claims.email)
  ```

- ✅ CORS plugin for both public and protected routes

## 🏗️ Architecture Flow

### Before (Old Approach)
```
Client → Kong → Auth-Service HTTP → Validate JWT → Business Logic
                                   ↑
                              Every request validates JWT
                              (Performance bottleneck)
```

### After (New Approach)
```
Client → Kong Gateway
         ↓
      JWT Plugin (validate locally)
         ↓
      Request Transformer (inject x-user-id)
         ↓
      gRPC-Gateway
         ↓
      Auth-Service gRPC → Auth Interceptor → Business Logic
                          ↓
                    Read x-user-id from metadata
                    (No JWT validation needed)
```

## 🔑 Key Benefits

1. **Performance**
   - Kong validates JWT với shared secret (không gọi auth-service)
   - Auth-service không phải validate mỗi request
   - Giảm network calls, giảm latency

2. **Scalability**
   - Auth-service có thể scale độc lập
   - Kong cache JWT public key/secret
   - Stateless JWT validation

3. **Security**
   - Centralized authentication tại gateway
   - Consistent JWT validation across services
   - Easy to add new microservices (không cần implement JWT validation)

4. **Maintainability**
   - Auth logic tách biệt khỏi business logic
   - Kong config declarative (GitOps friendly)
   - Easy to debug (logs tập trung tại gateway)

## 📝 Important Notes

### JWT Secret Synchronization
⚠️ **CRITICAL**: JWT secret phải giống nhau:
- Auth-service `.env`: `JWT_SECRET=your-secret-key-change-this-in-production`
- Kong `kong.yml`: `secret: your-secret-key-change-this-in-production`

### Token Blacklist Handling
Current approach: **Option B - Check in Service**
- Kong validates JWT signature & expiry
- Auth-service checks blacklist khi xử lý business logic
- Trade-off: Short delay for revocation vs simplicity

Future enhancement: Implement Kong custom plugin to check Redis blacklist

### RefreshToken Flow
Changed from cookie-based to body-based:
- **Before**: RefreshToken in HTTP-only cookie
- **After**: RefreshToken in request body (gRPC friendly)

### Audit Logs
Kong forwards client info via headers:
- `x-forwarded-for` → Client IP
- `user-agent` → User agent string
- Interceptor extracts these for audit logs

## 🧪 Testing

See `TESTING_GUIDE.md` for detailed testing instructions.

Quick test:
```bash
# 1. Register
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"Pass1234"}'

# 2. Login
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"Pass1234"}'

# 3. Get user info (use token from login)
curl http://localhost:8000/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

## 🚀 Deployment

### Production Checklist
- [ ] Change JWT_SECRET to strong random value
- [ ] Update Kong consumer secret to match JWT_SECRET
- [ ] Enable HTTPS on Kong (TLS termination)
- [ ] Set CORS origins to production domains only
- [ ] Monitor Kong metrics (request rate, latency, errors)
- [ ] Setup log aggregation (Kong + Auth-service logs)
- [ ] Configure Redis for distributed blacklist (future)

## 📚 Next Steps

1. **Add User-Service**
   - Follow same pattern (gRPC only)
   - Add route to Kong with JWT plugin
   - Use interceptor to get user_id from metadata

2. **Implement Redis Blacklist**
   - Shared Redis for all services
   - Kong custom plugin for real-time revocation

3. **Add Rate Limiting**
   - Kong rate-limit plugin per consumer
   - Protect against brute force attacks

4. **Setup Monitoring**
   - Prometheus metrics from Kong
   - Grafana dashboard for API Gateway
   - Alert on high error rates

## 🐛 Known Issues / Limitations

1. **Token Revocation Delay**
   - Blacklisted tokens still valid at Kong level until checked by service
   - Acceptable for most use cases (short TTL = 15 mins)

2. **gRPC Error Mapping**
   - Some gRPC errors may not map perfectly to HTTP status codes
   - Test all error scenarios

3. **No Distributed Tracing Yet**
   - Add OpenTelemetry in future for request tracing across services

## 📞 Support

If you encounter issues:
1. Check logs: Kong logs (`docker logs kong`) and auth-service logs
2. Verify JWT secret synchronization
3. Test with curl commands from TESTING_GUIDE.md
4. Check Kong admin API: `http://localhost:8001/routes`, `/plugins`
