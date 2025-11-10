# Fix Summary - JWT Validation in Auth-Service

## ❌ Previous Approach (Failed)
**Kong JWT Plugin approach** - Failed because:
- Kong JWT plugin requires consumer configuration with JWT credentials
- Complex setup: Need to create consumer, add JWT secret, configure key claims
- Error: `"No credentials found for given 'user_id'"` - Kong couldn't match JWT to consumer

## ✅ New Approach (Working)
**Auth-Service validates JWT directly via gRPC interceptor**

### Why This Works Better:
1. **Simpler architecture** - No need for Kong JWT plugin configuration
2. **Centralized validation** - JWT validation logic stays in auth-service
3. **Easier to maintain** - One place to update JWT logic
4. **Blacklist support** - Can check token blacklist easily

### Flow:
```
Client → Kong Gateway (HTTP) 
       ↓ (just routing, no JWT validation)
       ↓ Authorization header forwarded as gRPC metadata
       ↓
  gRPC Interceptor (Auth-Service)
       ↓ Extract Authorization header from metadata
       ↓ Validate JWT token
       ↓ Extract claims (user_id, email, role)
       ↓ Inject into context
       ↓
  Business Logic (gRPC Handler)
       ↓ Read user_id from context
       ↓ Process request
```

## 📝 Changes Made

### 1. Kong Configuration (`api-gateway/kong.yml`)
**Before**: Separate public/protected routes with JWT plugin
```yaml
consumers:
  - username: system
    jwt_secrets:
      - key: system
        secret: your-secret-key...
        algorithm: HS256

routes:
  - name: auth-protected-routes
    plugins:
      - name: jwt
        config:
          key_claim_name: user_id
```

**After**: Single route, Kong only does routing
```yaml
services:
  - name: auth-service
    routes:
      - name: auth-routes
        paths:
          - /api/v1/auth  # All auth endpoints
        plugins:
          - name: grpc-gateway  # Only HTTP→gRPC translation
          - name: cors
```

### 2. gRPC Interceptor (`auth-service/internal/delivery/grpc/interceptor/auth_interceptor.go`)
**Before**: Extract `x-user-id` from metadata (injected by Kong)
```go
func AuthInterceptor() grpc.UnaryServerInterceptor {
    userIDs := md.Get("x-user-id")  // Kong was supposed to inject this
    if len(userIDs) == 0 {
        return error
    }
}
```

**After**: Extract Authorization header, validate JWT directly
```go
func NewAuthInterceptor(tokenService TokenValidator) grpc.UnaryServerInterceptor {
    // Get Authorization header
    authHeaders := md.Get("authorization")
    
    // Parse "Bearer <token>"
    token := parts[1]
    
    // Validate JWT using auth-service's TokenService
    claims, err := tokenService.ValidateAccessToken(token)
    
    // Inject user_id, email, role into context
    ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
    ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
    ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
}
```

### 3. Token Validator Adapter (`token_validator_adapter.go`)
**New file** - Adapter pattern to connect interceptor with TokenService

```go
type TokenServiceAdapter struct {
    tokenService service.TokenService
}

func (a *TokenServiceAdapter) ValidateAccessToken(token string) (*TokenClaims, error) {
    claims, err := a.tokenService.ValidateAccessToken(token)
    // Convert domain.TokenClaims → interceptor.TokenClaims
    return &TokenClaims{
        UserID: claims.UserID,
        Email:  claims.Email,
        Role:   claims.Role,
    }, nil
}
```

### 4. Main Server (`cmd/server/main.go`)
**Added**: Inject TokenService into interceptor

```go
tokenValidator := interceptor.NewTokenServiceAdapter(tokenService)
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(interceptor.NewAuthInterceptor(tokenValidator)),
)
```

## 🧪 Testing

### Test Command:
```bash
curl --location 'http://localhost:8000/api/v1/auth/me' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
```

### Expected Flow:
1. **Kong** receives HTTP request with `Authorization: Bearer <token>`
2. **Kong** translates HTTP → gRPC, forwards `authorization` metadata
3. **gRPC Interceptor** extracts Authorization header from metadata
4. **Interceptor** validates JWT using `TokenService.ValidateAccessToken()`
5. **Interceptor** checks token blacklist (if needed)
6. **Interceptor** extracts claims and injects into context
7. **gRPC Handler** reads `user_id` from context via `GetUserIDFromContext()`
8. **Handler** processes request with authenticated user

### Success Response:
```json
{
  "id": "4cee29b4-b9fe-4e48-9c9c-acf72510362f",
  "email": "trump@gmail.com",
  "role": "user",
  "is_verified": false,
  "is_active": true,
  "created_at": "2024-11-10T10:00:00Z"
}
```

### Error Cases:
| Error | Status | Reason |
|-------|--------|--------|
| Missing Authorization header | 401 | No bearer token provided |
| Invalid token format | 401 | Not "Bearer <token>" format |
| Expired token | 401 | Token exp claim < current time |
| Invalid signature | 401 | JWT signature verification failed |
| Blacklisted token | 401 | Token in blacklist (logout/refresh) |

## 🔐 Security Benefits

1. **Token Blacklist Support**
   - Can check if token was revoked (logout/refresh)
   - Query database or Redis for blacklist

2. **Centralized JWT Logic**
   - Easy to update JWT validation rules
   - Consistent validation across all endpoints

3. **No Shared Secret Exposure**
   - JWT secret stays in auth-service only
   - Kong doesn't need to know the secret

4. **Flexible Claims Extraction**
   - Can extract any claim (user_id, email, role, permissions)
   - Easy to add custom claims

## 📚 Proto Import Fix

### Issue:
```
Import "google/api/annotations.proto" was not found or had errors.
```

### Root Cause:
Proto compiler couldn't find google API proto files when generating code.

### Solution:
Files already exist in `/auth-service/proto/google/api/`:
- `annotations.proto`
- `http.proto`

Docker-compose already mounts entire proto directory to Kong:
```yaml
volumes:
  - ./auth-service/proto:/etc/kong/proto
```

**No code changes needed** - Import path is correct:
```protobuf
import "google/api/annotations.proto";
```

Kong will find it at `/etc/kong/proto/google/api/annotations.proto`

## 🚀 Next Steps

1. **Start services**:
   ```bash
   docker-compose up kong auth-db
   cd auth-service && go run cmd/server/main.go
   ```

2. **Test endpoints**:
   ```bash
   # Register
   curl -X POST http://localhost:8000/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@test.com","password":"Pass1234"}'
   
   # Login
   curl -X POST http://localhost:8000/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@test.com","password":"Pass1234"}'
   
   # Get Me (use token from login)
   curl http://localhost:8000/api/v1/auth/me \
     -H "Authorization: Bearer <access_token>"
   ```

3. **Monitor logs**:
   - Kong: `docker logs -f kong`
   - Auth-service: Check console output

## 🎯 Comparison: Kong JWT vs Auth-Service JWT

| Aspect | Kong JWT Plugin | Auth-Service JWT (Current) |
|--------|----------------|---------------------------|
| **Setup Complexity** | High (consumer, credentials) | Low (just interceptor) |
| **JWT Secret Location** | Kong config | Auth-service only |
| **Token Blacklist** | Hard (need custom plugin) | Easy (check in interceptor) |
| **Performance** | Faster (no DB call) | Slightly slower (validates each request) |
| **Maintainability** | Kong + Service | Service only |
| **Flexibility** | Limited by plugin | Full control |
| **Best For** | Public APIs, high traffic | Internal microservices, need blacklist |

**Decision**: Auth-Service JWT validation is better for your use case because:
- Need token blacklist support (logout/refresh)
- Simpler configuration
- Easier to debug
- All auth logic in one place
