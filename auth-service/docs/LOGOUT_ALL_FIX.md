# LogoutAll Security Fix

## Vấn đề ban đầu

User phát hiện ra lỗ hổng bảo mật quan trọng trong hàm `LogoutAll`:

```go
// TRƯỚC - Có lỗ hổng
func (uc *AuthUseCase) LogoutAll(ctx context.Context, userID string, ...) error {
    // Chỉ revoke refresh tokens
    uc.refreshTokenRepo.RevokeAllByUserID(ctx, userUUID)
    // ❌ Access tokens vẫn hoạt động cho đến khi hết hạn (15 phút)
}
```

**Vấn đề:**
- User gọi `LogoutAll` (ví dụ: logout khỏi tất cả devices)
- Tất cả refresh tokens bị revoke ✅
- Nhưng access tokens hiện tại vẫn valid ❌
- Trong 15 phút đó, attacker/device khác vẫn gọi được API

**Tại sao nghiêm trọng?**
- LogoutAll thường dùng khi:
  - Phát hiện account bị compromise
  - Mất thiết bị
  - Muốn force logout tất cả sessions
- Nếu access token vẫn dùng được → mất tác dụng của LogoutAll!

---

## Giải pháp

### Solution: User-level Token Invalidation

Thêm timestamp `LastLogoutAllAt` vào User entity. Khi LogoutAll:
1. Set `user.LastLogoutAllAt = now()`
2. Middleware check: if `token.IssuedAt < user.LastLogoutAllAt` → reject

**Ưu điểm:**
- Đơn giản, dễ hiểu
- Không cần lưu tất cả access tokens (stateless JWT)
- Hiệu quả: 1 query extra trong middleware (có thể cache)

---

## Implementation

### 1. User Entity - Thêm LastLogoutAllAt

```go
// internal/domain/entity/user.go

type User struct {
    // ... existing fields
    LastLogoutAllAt *time.Time  // ✅ NEW
    // ...
}

func (u *User) LogoutAll() {
    now := time.Now()
    u.LastLogoutAllAt = &now
    u.UpdatedAt = now
}
```

### 2. Database Model

```go
// internal/infrastructure/persistence/postgres/user_repository.go

type UserModel struct {
    // ... existing fields
    LastLogoutAllAt *time.Time  // ✅ NEW
    // ...
}

// Update toModel() và toEntity() để map field này
```

### 3. TokenClaims - Thêm IssuedAt

```go
// internal/domain/service/token_service.go

type TokenClaims struct {
    UserID   string
    Email    string
    Role     string
    IssuedAt int64  // ✅ NEW - Unix timestamp
}
```

### 4. JWT Service - Return IssuedAt

```go
// internal/infrastructure/security/jwt_service.go

func (s *JWTService) ValidateAccessToken(tokenString string) (*service.TokenClaims, error) {
    // ... parse token
    
    var issuedAt int64
    if claims.IssuedAt != nil {
        issuedAt = claims.IssuedAt.Unix()  // ✅ Extract from JWT
    }
    
    return &service.TokenClaims{
        UserID:   claims.UserID,
        Email:    claims.Email,
        Role:     claims.Role,
        IssuedAt: issuedAt,  // ✅ NEW
    }, nil
}
```

### 5. LogoutAll UseCase - Update User

```go
// internal/application/usecase/auth_usecase.go

func (uc *AuthUseCase) LogoutAll(ctx context.Context, userID string, ...) error {
    userUUID, _ := uuid.Parse(userID)
    
    // ✅ Fetch user
    user, err := uc.userRepo.FindByID(ctx, userUUID)
    if err != nil {
        return domainErr.ErrUserNotFound
    }
    
    // ✅ Set LastLogoutAllAt timestamp
    user.LogoutAll()
    
    // ✅ Save to database
    if err := uc.userRepo.Update(ctx, user); err != nil {
        return domainErr.ErrDatabase
    }
    
    // Revoke all refresh tokens (existing)
    uc.refreshTokenRepo.RevokeAllByUserID(ctx, userUUID)
    
    // ... audit log
    return nil
}
```

### 6. Middleware - Check LastLogoutAllAt

```go
// internal/delivery/http/middleware/auth_middleware.go

type AuthMiddleware struct {
    tokenService       service.TokenService
    tokenBlacklistRepo repository.TokenBlacklistRepository
    userRepo           repository.UserRepository  // ✅ NEW
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... validate token, check blacklist (existing)
        
        // ✅ NEW: Check LastLogoutAllAt
        userID, err := uuid.Parse(claims.UserID)
        if err == nil {
            user, err := m.userRepo.FindByID(c.Request.Context(), userID)
            if err == nil && user.LastLogoutAllAt != nil {
                tokenIssuedAt := time.Unix(claims.IssuedAt, 0)
                
                // If token issued before last LogoutAll → reject
                if tokenIssuedAt.Before(*user.LastLogoutAllAt) {
                    c.JSON(http.StatusUnauthorized, gin.H{
                        "error": "token has been revoked",
                    })
                    c.Abort()
                    return
                }
            }
        }
        
        // ... continue
    }
}
```

---

## Flow hoạt động

### Normal Flow (Không có LogoutAll)

```
1. User login
   → Token issued at: 2025-01-01 10:00:00
   → user.LastLogoutAllAt = nil

2. API Request
   → Middleware checks:
     ✅ Token valid
     ✅ Not blacklisted
     ✅ user.LastLogoutAllAt = nil → OK
   → Request processed
```

### After LogoutAll

```
1. User calls LogoutAll at 2025-01-01 10:30:00
   → user.LastLogoutAllAt = 2025-01-01 10:30:00
   → All refresh tokens revoked

2. Attacker tries to use old access token (issued at 10:00:00)
   → Middleware checks:
     ✅ Token signature valid
     ✅ Token not expired
     ✅ Not in blacklist
     ❌ token.IssuedAt (10:00) < user.LastLogoutAllAt (10:30)
   → Request REJECTED with 401

3. User logs in again at 10:35:00
   → New token issued at: 2025-01-01 10:35:00
   → token.IssuedAt (10:35) > user.LastLogoutAllAt (10:30)
   → ✅ Works normally
```

---

## Database Migration

Auto-migration sẽ tự động thêm column:

```sql
ALTER TABLE users 
ADD COLUMN last_logout_all_at TIMESTAMP NULL;
```

Không cần migration script riêng vì GORM AutoMigrate sẽ xử lý.

---

## Performance Considerations

### Extra Query per Request?

**Có**, middleware giờ query user:

```go
user, err := m.userRepo.FindByID(c.Request.Context(), userID)
```

**Impact:**
- 1 query SELECT vào users table mỗi request
- Có index trên primary key (ID) → rất nhanh
- Điển hình: <1ms với PostgreSQL

**Optimization options (future):**

1. **Cache user.LastLogoutAllAt** (Redis)
```go
// Check cache first
lastLogoutAllAt, err := redis.Get("user:"+userID+":last_logout_all")
if err == nil {
    // Use cached value
} else {
    // Fallback to DB
}
```

2. **Only check if LastLogoutAllAt might exist**
```go
// Skip check for users who never called LogoutAll
// Can track in separate table or Redis set
```

3. **Batch/Pipeline queries**
```go
// If multiple auth checks, batch them
```

**Recommendation:**
- Start without cache (simpler)
- Monitor performance
- Add cache if needed (typically not needed for <1000 req/s)

---

## Testing

### Test 1: Normal flow (no LogoutAll)
```bash
# Login
TOKEN=$(curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.access_token')

# Use token (should work)
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
# ✅ Success
```

### Test 2: LogoutAll invalidates access token
```bash
# Login
TOKEN=$(curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.access_token')

# Use token (works)
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
# ✅ Success

# LogoutAll
curl -X POST http://localhost:9001/api/v1/auth/logout-all \
  -H "Authorization: Bearer $TOKEN"
# ✅ Success

# Try to use same token (should fail)
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
# ❌ 401 Unauthorized: token has been revoked
```

### Test 3: New login works after LogoutAll
```bash
# LogoutAll (from previous test)

# Login again
NEW_TOKEN=$(curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.access_token')

# Use new token (should work)
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer $NEW_TOKEN"
# ✅ Success
```

---

## Security Benefits

### Before Fix ❌
```
Scenario: Account compromised, user clicks "Logout all devices"

Attacker:
- Old access token still valid for 15 minutes
- Can continue accessing API
- Can read/modify user data
- User thinks they're safe but they're not!
```

### After Fix ✅
```
Scenario: Account compromised, user clicks "Logout all devices"

Attacker:
- Old access token immediately invalid
- Cannot access any API endpoints
- User is actually safe

User:
- Needs to login again
- Gets new token with timestamp > LastLogoutAllAt
- Everything works normally
```

---

## Alternative Solutions (Not chosen)

### ❌ Option 1: Store all access tokens
```go
// Store every access token in database
// When LogoutAll → blacklist all of them
```
**Problems:**
- Defeats purpose of stateless JWT
- Huge database growth
- Performance hit on every token generation

### ❌ Option 2: Short-lived access tokens
```go
// Make access token expire in 1 minute
```
**Problems:**
- Too many refresh requests
- Bad UX
- Doesn't solve the core issue

### ✅ Option 3: LastLogoutAllAt (CHOSEN)
**Why best:**
- Simple implementation
- Minimal performance impact
- Solves the problem completely
- Can be cached if needed

---

## Summary

**Changes:**
- ✅ User entity: +1 field (LastLogoutAllAt)
- ✅ TokenClaims: +1 field (IssuedAt)
- ✅ LogoutAll: Update user timestamp
- ✅ Middleware: +1 check (token issued vs last logout)

**Security improvement:**
- Before: Access tokens valid for 15 min after LogoutAll ❌
- After: Access tokens immediately invalid ✅

**Performance:**
- +1 DB query per request (user lookup)
- Can be optimized with cache if needed
- Acceptable trade-off for security

**Build:** ✅ Successful
