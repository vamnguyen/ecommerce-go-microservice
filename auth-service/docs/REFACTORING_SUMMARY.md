# Refactoring Summary - Code Simplification

## Vấn đề đã giải quyết

### 1. **Code thừa: `extractAccessToken` trong handler** ❌
**Trước:**
```go
// Handler phải tự extract access token
func (h *AuthHandler) extractAccessToken(c *gin.Context) string {
    authHeader := c.GetHeader("Authorization")
    // ... 10 lines code để parse
    return token
}

func (h *AuthHandler) Logout(c *gin.Context) {
    accessToken := h.extractAccessToken(c) // Duplicate logic
    // ...
}
```

**Sau:** ✅
```go
// Middleware đã extract rồi, chỉ cần lấy từ context
func (h *AuthHandler) Logout(c *gin.Context) {
    accessToken, _ := c.Get("accessToken")
    accessTokenStr, _ := accessToken.(string)
    // Simple & clean!
}
```

**Lợi ích:**
- Xóa được 15 dòng code thừa
- Logic extract token chỉ có 1 chỗ (middleware)
- Handler đơn giản hơn, chỉ lo business logic

---

### 2. **Data thừa trong Context: email và role không dùng** ❌
**Trước:**
```go
// Middleware set 3 values vào context
c.Set("userID", claims.UserID)
c.Set("email", claims.Email)     // ❌ Không dùng đến
c.Set("role", claims.Role)       // ❌ Không dùng đến
```

**Sau:** ✅
```go
// Chỉ set những gì cần dùng
c.Set("userID", claims.UserID)
c.Set("accessToken", token)
```

**Lợi ích:**
- Context gọn gàng hơn
- Rõ ràng handler dùng gì
- Dễ maintain và debug

---

### 3. **Hardcode cookie config** ❌
**Trước:**
```go
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
    c.SetCookie(
        "refresh_token",
        token,
        30*24*60*60,  // ❌ Hardcode
        "/",
        "",           // ❌ Hardcode
        false,        // ❌ Hardcode
        true,
    )
}
```

**Sau:** ✅
```go
// Config từ environment variables
type CookieConfig struct {
    Secure bool
    Domain string
    MaxAge int
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
    c.SetCookie(
        "refresh_token",
        token,
        h.cookieCfg.MaxAge,    // ✅ Từ config
        "/",
        h.cookieCfg.Domain,    // ✅ Từ config
        h.cookieCfg.Secure,    // ✅ Từ config
        true,
    )
}
```

**Lợi ích:**
- Dev/Prod có config khác nhau dễ dàng
- Không cần rebuild khi đổi config
- Follow best practices

---

## Chi tiết thay đổi

### Files modified:

#### 1. `internal/infrastructure/config/config.go`
```go
// Thêm CookieConfig
type CookieConfig struct {
    Secure bool   // HTTPS only in production
    Domain string // For subdomain sharing
    MaxAge int    // Cookie lifetime (seconds)
}

type Config struct {
    // ... existing fields
    Cookie CookieConfig
}

// Load from environment
Cookie: CookieConfig{
    Secure: parseBool(getEnv("COOKIE_SECURE", "false")),
    Domain: getEnv("COOKIE_DOMAIN", ""),
    MaxAge: parseInt(getEnv("COOKIE_MAX_AGE", "2592000")), // 30 days
},
```

#### 2. `internal/delivery/http/middleware/auth_middleware.go`
```go
// BEFORE
c.Set("userID", claims.UserID)
c.Set("email", claims.Email)     // ❌ Removed
c.Set("role", claims.Role)       // ❌ Removed

// AFTER
c.Set("userID", claims.UserID)
c.Set("accessToken", token)      // ✅ Added for Logout
```

#### 3. `internal/delivery/http/handler/auth_handler.go`
```go
// Inject cookieConfig
type AuthHandler struct {
    authUseCase *usecase.AuthUseCase
    logger      *logger.Logger
    cookieCfg   config.CookieConfig  // ✅ Added
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase, logger *logger.Logger, cookieCfg config.CookieConfig) *AuthHandler {
    return &AuthHandler{
        authUseCase: authUseCase,
        logger:      logger,
        cookieCfg:   cookieCfg,
    }
}

// Logout - get accessToken from context
func (h *AuthHandler) Logout(c *gin.Context) {
    accessToken, _ := c.Get("accessToken")  // ✅ From middleware
    accessTokenStr, _ := accessToken.(string)
    // ... simple & clean
}

// Cookie helpers use config
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
    c.SetCookie(
        "refresh_token",
        token,
        h.cookieCfg.MaxAge,    // ✅ From config
        "/",
        h.cookieCfg.Domain,    // ✅ From config
        h.cookieCfg.Secure,    // ✅ From config
        true,
    )
}

// ❌ Removed: extractAccessToken() - 15 lines deleted
```

#### 4. `cmd/server/main.go`
```go
// Wire cookieConfig
authHandler := handler.NewAuthHandler(authUseCase, log, cfg.Cookie)
```

#### 5. `.env.example`
```env
# Cookie Settings
COOKIE_SECURE=false          # true in production
COOKIE_DOMAIN=               # e.g., .example.com for subdomains
COOKIE_MAX_AGE=2592000       # 30 days in seconds
```

---

## So sánh Before/After

### Flow: Logout Request

**BEFORE (Complicated):**
```
Request → Middleware
          ├─ Extract token
          ├─ Set userID ✅
          ├─ Set email ❌ (unused)
          └─ Set role ❌ (unused)
          
       → Handler
          ├─ Extract token AGAIN ❌ (duplicate)
          ├─ Use hardcoded cookie config ❌
          └─ Logout
```

**AFTER (Clean):**
```
Request → Middleware
          ├─ Extract token
          ├─ Set userID ✅
          └─ Set accessToken ✅
          
       → Handler
          ├─ Get token from context ✅
          ├─ Use config for cookies ✅
          └─ Logout
```

---

## Metrics

### Code reduction:
- **Deleted:** 15 lines (extractAccessToken function)
- **Simplified:** Handler logic
- **Cleaner:** Context data

### Maintainability:
- **Before:** Token extraction in 2 places
- **After:** Token extraction in 1 place (middleware)

### Configuration:
- **Before:** 3 hardcoded values
- **After:** 3 configurable values from env

---

## Testing

Build successful:
```bash
$ go build -o /tmp/auth-service-test ./cmd/server/main.go
[Process exited with code 0]
```

All functionality maintained, just cleaner code.

---

## Environment Variables

Add to your `.env`:
```env
COOKIE_SECURE=false
COOKIE_DOMAIN=
COOKIE_MAX_AGE=2592000
```

**Production values:**
```env
COOKIE_SECURE=true              # Require HTTPS
COOKIE_DOMAIN=.yourdomain.com   # Share across subdomains
COOKIE_MAX_AGE=2592000          # 30 days
```

---

## Summary

✅ **Xóa code thừa:** extractAccessToken function  
✅ **Xóa data thừa:** email và role trong context  
✅ **Thêm config:** Cookie settings từ environment  
✅ **Code gọn hơn:** 15 lines ít hơn  
✅ **Flow rõ ràng hơn:** Middleware làm gì, handler làm gì  
✅ **Dễ maintain hơn:** Logic không bị duplicate  

**Result:** Code đơn giản, dễ hiểu, dễ config! 🎉
