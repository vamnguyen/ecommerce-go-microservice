# Token Security Migration Guide

## Overview
This migration adds enhanced token security features:
1. **Cookie-based Refresh Token** - Refresh tokens are stored in HTTP-only cookies
2. **Token Family/Rotation** - Detects token reuse and revokes entire token family
3. **Access Token Blacklist** - Revoked refresh tokens also blacklist their access tokens

## Database Changes

### 1. Add token_family_id to refresh_tokens table

```sql
-- Add token_family_id column to refresh_tokens
ALTER TABLE refresh_tokens 
ADD COLUMN token_family_id UUID NOT NULL DEFAULT gen_random_uuid();

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_family_id 
ON refresh_tokens(token_family_id);
```

### 2. Create token_blacklist table

```sql
-- Create token_blacklist table
CREATE TABLE IF NOT EXISTS token_blacklist (
    id UUID PRIMARY KEY,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_token_blacklist_token_hash 
ON token_blacklist(token_hash);

CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at 
ON token_blacklist(expires_at);
```

## Auto Migration

The application will automatically create these tables and columns when it starts up using GORM's AutoMigrate feature.

## Breaking Changes

### API Changes

#### 1. Login Response
**Before:**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "abc123...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {...}
}
```

**After:**
```json
{
  "access_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {...}
}
```
- `refresh_token` is now sent as HTTP-only cookie instead of response body

#### 2. Refresh Token Endpoint

**Before:**
```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "abc123..."
}
```

**After:**
```bash
POST /api/v1/auth/refresh
Cookie: refresh_token=abc123...

# No request body needed
```

**Response Before:**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "xyz789...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Response After:**
```json
{
  "access_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900
}
```
- New refresh token is sent as HTTP-only cookie

#### 3. Logout Endpoint

**Before:**
```bash
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGc...
Content-Type: application/json

{
  "refresh_token": "abc123..."
}
```

**After:**
```bash
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGc...
Cookie: refresh_token=abc123...

# No request body needed
```
- Both access token and refresh token are now blacklisted

## Security Improvements

### 1. Token Family/Rotation
- Each refresh token belongs to a token family (identified by `token_family_id`)
- When a refresh token is used, it's revoked and a new one is issued in the same family
- If a revoked token is reused (possible attack), the entire family is revoked
- This protects against token theft and replay attacks

### 2. Access Token Blacklist
- When logout occurs, the access token is added to a blacklist
- Middleware checks the blacklist before accepting tokens
- Expired tokens are automatically cleaned from the blacklist
- This prevents logout bypass by continuing to use valid access tokens

### 3. Cookie Security
- Refresh tokens are stored in HTTP-only cookies (not accessible via JavaScript)
- Protects against XSS attacks
- Cookies are automatically sent with each request
- No need for client-side token management

## Environment Variables

Add these new variables to your `.env` file:

```env
# Cookie Settings
COOKIE_SECURE=false        # Set to true in production (requires HTTPS)
COOKIE_DOMAIN=             # Optional: set for subdomain sharing
```

## Client-Side Changes Required

### JavaScript/TypeScript Example

**Before:**
```typescript
// Login
const response = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email, password })
});
const { access_token, refresh_token } = await response.json();
localStorage.setItem('access_token', access_token);
localStorage.setItem('refresh_token', refresh_token);

// Refresh
const refreshResponse = await fetch('/api/v1/auth/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ refresh_token: localStorage.getItem('refresh_token') })
});
```

**After:**
```typescript
// Login
const response = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  credentials: 'include',  // Important: include cookies
  body: JSON.stringify({ email, password })
});
const { access_token } = await response.json();
localStorage.setItem('access_token', access_token);
// No need to store refresh_token - it's in cookie

// Refresh
const refreshResponse = await fetch('/api/v1/auth/refresh', {
  method: 'POST',
  credentials: 'include'  // Cookie sent automatically
});
```

**Key Changes:**
1. Add `credentials: 'include'` to all fetch requests
2. Don't store refresh token in localStorage
3. Don't send refresh token in request body

## Testing

### 1. Test Cookie-based Refresh Token
```bash
# Login
curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  -c cookies.txt

# Refresh (cookie sent automatically)
curl -X POST http://localhost:9001/api/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```

### 2. Test Token Reuse Detection
```bash
# 1. Get refresh token
curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  -c cookies1.txt

# 2. Copy cookie for testing
cp cookies1.txt cookies2.txt

# 3. Use first cookie (should work)
curl -X POST http://localhost:9001/api/v1/auth/refresh \
  -b cookies1.txt \
  -c cookies1.txt

# 4. Try to use old cookie (should fail and revoke family)
curl -X POST http://localhost:9001/api/v1/auth/refresh \
  -b cookies2.txt
# Expected: 401 Unauthorized - token revoked
```

### 3. Test Access Token Blacklist
```bash
# 1. Login and get access token
RESPONSE=$(curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}')
ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.access_token')

# 2. Access protected endpoint (should work)
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. Logout
curl -X POST http://localhost:9001/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 4. Try to use same access token (should fail)
curl -X GET http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# Expected: 401 Unauthorized - token has been revoked
```

## Rollback Plan

If you need to rollback:

```sql
-- Remove token_family_id column
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS token_family_id;

-- Drop token_blacklist table
DROP TABLE IF EXISTS token_blacklist;

-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_token_family_id;
DROP INDEX IF EXISTS idx_token_blacklist_token_hash;
DROP INDEX IF EXISTS idx_token_blacklist_expires_at;
```

## Cleanup Jobs

Consider adding cron jobs to clean up expired tokens:

```go
// Example cleanup function (to be scheduled)
func CleanupExpiredTokens(ctx context.Context) {
    // Clean expired refresh tokens
    refreshTokenRepo.DeleteExpired(ctx)
    
    // Clean expired blacklisted tokens
    tokenBlacklistRepo.DeleteExpired(ctx)
}
```

Recommended schedule: Every 24 hours
