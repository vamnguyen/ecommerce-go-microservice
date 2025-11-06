# Token Security Implementation Summary

## Tổng quan

Đã implement 3 tính năng bảo mật quan trọng cho hệ thống authentication:

1. **Cookie-based Refresh Token** - Refresh token lưu trong HTTP-only cookie
2. **Token Family/Rotation** - Phát hiện token reuse và revoke toàn bộ family
3. **Access Token Blacklist** - Khi revoke refresh token thì access token cũng bị blacklist

## Vấn đề đã giải quyết

### Vấn đề 1: Refresh Token trong Request Body (Không an toàn)
**Trước:**
- Refresh token gửi qua request body
- Client phải lưu trong localStorage (dễ bị XSS attack)
- Dễ bị đánh cắp qua JavaScript malicious

**Sau:**
- Refresh token lưu trong HTTP-only cookie
- JavaScript không thể truy cập (chống XSS)
- Tự động gửi với mọi request đến server

### Vấn đề 2: Token Reuse không được phát hiện
**Trước:**
- Nếu refresh token bị đánh cắp, attacker có thể dùng mãi
- Không có cơ chế phát hiện token bị dùng lại

**Sau:**
- Mỗi refresh token thuộc một token family (có token_family_id)
- Khi refresh, token cũ bị revoke, token mới vẫn cùng family
- Nếu token đã revoke bị dùng lại → toàn bộ family bị revoke
- User thật phải login lại, attacker không dùng được nữa

### Vấn đề 3: Access Token vẫn hoạt động sau khi logout
**Trước:**
- Logout chỉ revoke refresh token
- Access token vẫn valid cho đến khi hết hạn (15 phút)
- Có thể tiếp tục gọi API trong thời gian này

**Sau:**
- Khi logout, cả access token và refresh token đều bị revoke
- Access token thêm vào blacklist table
- Middleware check blacklist trước khi cho phép request
- Access token không còn hoạt động ngay lập tức

## Các file đã thay đổi

### 1. Domain Layer

#### Entities
- ✅ `internal/domain/entity/refresh_token.go` - Thêm `TokenFamilyID` field và `NewRefreshTokenWithFamily` method
- ✅ `internal/domain/entity/token_blacklist.go` - Entity mới cho blacklist

#### Repositories
- ✅ `internal/domain/repository/refresh_token_repository.go` - Thêm method `RevokeByTokenFamilyID`
- ✅ `internal/domain/repository/token_blacklist_repository.go` - Repository interface mới

### 2. Infrastructure Layer

#### Persistence
- ✅ `internal/infrastructure/persistence/postgres/refresh_token_repository.go`
  - Thêm `TokenFamilyID` field vào model
  - Implement `RevokeByTokenFamilyID` method
  
- ✅ `internal/infrastructure/persistence/postgres/token_blacklist_repository.go` - Implementation mới
  - `Add()` - Thêm token vào blacklist
  - `IsBlacklisted()` - Check token có bị blacklist không
  - `DeleteExpired()` - Xóa token đã hết hạn

- ✅ `internal/infrastructure/persistence/postgres/database.go` - Thêm `TokenBlacklistModel` vào AutoMigrate

### 3. Application Layer

#### Use Cases
- ✅ `internal/application/usecase/auth_usecase.go`
  - Thêm `tokenBlacklistRepo` dependency
  - Update `RefreshToken()` method:
    - Phát hiện token reuse và revoke family
    - Sử dụng `NewRefreshTokenWithFamily` để giữ family ID
  - Update `Logout()` method:
    - Thêm parameter `accessToken`
    - Blacklist access token khi logout

### 4. Delivery Layer

#### Handlers
- ✅ `internal/delivery/http/handler/auth_handler.go`
  - Update `Login()`: Set refresh token vào cookie, không trả trong response
  - Update `RefreshToken()`: Đọc từ cookie thay vì request body
  - Update `Logout()`: Đọc refresh token từ cookie và extract access token từ header
  - Thêm helper methods:
    - `setRefreshTokenCookie()` - Set HTTP-only cookie
    - `clearRefreshTokenCookie()` - Clear cookie khi logout
    - `extractAccessToken()` - Extract token từ Authorization header

#### Middleware
- ✅ `internal/delivery/http/middleware/auth_middleware.go`
  - Thêm `tokenBlacklistRepo` dependency
  - Update `RequireAuth()`: Check token có trong blacklist không trước khi accept

### 5. Main Application
- ✅ `cmd/server/main.go`
  - Initialize `tokenBlacklistRepo`
  - Wire vào `authUseCase` và `authMiddleware`

### 6. Configuration
- ✅ `.env.example` - Thêm cookie settings:
  - `COOKIE_SECURE` - Set true trong production
  - `COOKIE_DOMAIN` - Cho subdomain sharing

### 7. Documentation
- ✅ `docs/MIGRATION_TOKEN_SECURITY.md` - Migration guide chi tiết
- ✅ `docs/TOKEN_SECURITY_IMPLEMENTATION.md` - Tài liệu này

## Database Schema Changes

### refresh_tokens table
```sql
-- Thêm column mới
token_family_id UUID NOT NULL
```

### token_blacklist table (mới)
```sql
CREATE TABLE token_blacklist (
    id UUID PRIMARY KEY,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

## Flow hoạt động

### 1. Login Flow
```
Client → POST /api/v1/auth/login
         ↓
Server:  - Verify credentials
         - Generate access token + refresh token
         - Create refresh token entity with new family ID
         - Save to database
         - Set refresh token in HTTP-only cookie
         - Return access token in response body
         ↓
Client:  - Store access token in memory/localStorage
         - Browser automatically stores cookie
```

### 2. Refresh Token Flow (Normal Case)
```
Client → POST /api/v1/auth/refresh (với cookie)
         ↓
Server:  - Read refresh token from cookie
         - Validate token
         - Check if already revoked (not revoked = OK)
         - Revoke old token
         - Generate new tokens (same family)
         - Set new refresh token in cookie
         - Return new access token
         ↓
Client:  - Update access token
         - Cookie updated automatically
```

### 3. Refresh Token Flow (Token Reuse Detected)
```
Attacker → POST /api/v1/auth/refresh (với stolen token đã revoked)
           ↓
Server:    - Read refresh token from cookie
           - Validate token
           - Detect token is revoked
           - 🚨 SECURITY ALERT: Token reuse detected!
           - Revoke entire token family
           - Return 401 Unauthorized
           ↓
User:      - Next request fails (family revoked)
           - Must login again
Attacker:  - Cannot use any token from that family
```

### 4. Logout Flow
```
Client → POST /api/v1/auth/logout
         - Authorization: Bearer <access_token>
         - Cookie: refresh_token=<refresh_token>
         ↓
Server:  - Extract access token from header
         - Read refresh token from cookie
         - Revoke refresh token in database
         - Add access token to blacklist
         - Clear refresh token cookie
         - Return 204 No Content
         ↓
Client:  - Cookie cleared
         - Access token immediately invalid
```

### 5. Protected Route with Blacklist Check
```
Client → GET /api/v1/auth/me
         - Authorization: Bearer <access_token>
         ↓
Middleware:
         - Extract token from header
         - Validate token signature & expiry
         - Hash token
         - Check if in blacklist ← NEW!
         - If blacklisted → 401 Unauthorized
         - If valid → Continue to handler
         ↓
Handler: - Process request
```

## Testing

### Manual Testing với curl

```bash
# 1. Login và lưu cookie
curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  -c cookies.txt -v

# 2. Access protected route
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer <access_token_from_step_1>"

# 3. Refresh token (cookie tự động gửi)
curl -X POST http://localhost:9001/api/v1/auth/refresh \
  -b cookies.txt -c cookies.txt -v

# 4. Logout
curl -X POST http://localhost:9001/api/v1/auth/logout \
  -H "Authorization: Bearer <access_token>" \
  -b cookies.txt -v

# 5. Try access with blacklisted token (should fail)
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer <access_token_from_step_4>" -v
```

## Lợi ích Bảo mật

### 1. Chống XSS (Cross-Site Scripting)
- Refresh token trong HTTP-only cookie
- JavaScript không thể đọc được
- Kể cả khi inject malicious script vào page

### 2. Chống CSRF (Cross-Site Request Forgery)
- Access token vẫn gửi qua header (không tự động)
- Attacker không thể forge request với access token

### 3. Chống Token Theft
- Token family detection: Nếu token bị stolen và reused → revoke toàn bộ
- User thật sẽ bị logout nhưng attacker cũng không dùng được
- User nhận ra bị attack khi bị logout bất thường

### 4. Immediate Token Revocation
- Không phải đợi access token expire
- Logout có hiệu lực ngay lập tức
- Quan trọng cho security incidents

## Performance Considerations

### 1. Blacklist Check
- Mỗi request đều check blacklist → có thể slow
- **Giải pháp**: 
  - Index trên `token_hash` column (đã implement)
  - Có thể dùng Redis cache cho blacklist (future improvement)
  - Auto cleanup expired tokens

### 2. Database Size
- Blacklist table có thể lớn nếu nhiều logout
- **Giải pháp**:
  - Chỉ lưu đến khi token expire
  - Cronjob cleanup expired tokens
  - Consider TTL index (PostgreSQL/MongoDB)

### 3. Cookie Size
- Cookie gửi với mọi request → bandwidth
- **Impact**: Minimal (refresh token ~50-100 bytes)

## Future Improvements

### 1. Redis Cache cho Blacklist
```go
// Check Redis first, fallback to DB
func (r *TokenBlacklistRepository) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
    // Check Redis cache
    cached, err := r.redis.Get(ctx, "blacklist:"+tokenHash).Result()
    if err == nil {
        return cached == "1", nil
    }
    
    // Fallback to database
    return r.isBlacklistedFromDB(ctx, tokenHash)
}
```

### 2. Rate Limiting cho Refresh Endpoint
- Prevent brute force token guessing
- Implement trong middleware

### 3. Device Tracking
- Thêm device_id vào refresh token
- User có thể xem và revoke từng device
- Detect suspicious device changes

### 4. Notification on Token Reuse
- Email/SMS alert khi detect token reuse
- User nhận biết account bị compromise

## Maintenance

### Cleanup Job
Nên chạy định kỳ để cleanup expired tokens:

```go
// Example cron job
func CleanupExpiredTokens(ctx context.Context, 
    refreshTokenRepo repository.RefreshTokenRepository,
    blacklistRepo repository.TokenBlacklistRepository) {
    
    log.Info("Starting token cleanup...")
    
    // Clean expired refresh tokens
    if err := refreshTokenRepo.DeleteExpired(ctx); err != nil {
        log.Error("Failed to cleanup refresh tokens", zap.Error(err))
    }
    
    // Clean expired blacklisted tokens
    if err := blacklistRepo.DeleteExpired(ctx); err != nil {
        log.Error("Failed to cleanup blacklist", zap.Error(err))
    }
    
    log.Info("Token cleanup completed")
}

// Schedule: Every 24 hours at 3 AM
// Cron: 0 3 * * *
```

## References

- [OWASP Token Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [RFC 6819 - OAuth 2.0 Threat Model](https://datatracker.ietf.org/doc/html/rfc6819)
- [Token Rotation Best Practices](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation)
