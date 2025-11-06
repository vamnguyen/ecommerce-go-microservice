# Auth Service API Documentation

Base URL: `http://localhost:8080`

## Authentication

Most endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <access_token>
```

## Endpoints

### 1. Health Check

Check if the service is running.

**Endpoint:** `GET /health`

**Authentication:** None

**Response:**
```json
{
  "status": "ok",
  "service": "auth-service"
}
```

---

### 2. Register User

Create a new user account.

**Endpoint:** `POST /api/v1/auth/register`

**Authentication:** None

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "StrongP@ss123"
}
```

**Password Requirements:**
- Minimum 8 characters
- Maximum 128 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character
- Cannot be common passwords (password, 12345678, etc.)

**Success Response (201 Created):**
```json
{
  "message": "user registered successfully"
}
```

**Error Responses:**

400 Bad Request - Invalid input:
```json
{
  "error": "invalid request payload"
}
```

400 Bad Request - Weak password:
```json
{
  "error": "password is too weak"
}
```

409 Conflict - User already exists:
```json
{
  "error": "user already exists"
}
```

---

### 3. Login

Authenticate user and receive tokens.

**Endpoint:** `POST /api/v1/auth/login`

**Authentication:** None

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "StrongP@ss123"
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "user",
    "is_verified": false,
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Response Fields:**
- `access_token`: JWT token for API access (expires in 15 minutes)
- `refresh_token`: Token to get new access tokens (expires in 30 days)
- `token_type`: Always "Bearer"
- `expires_in`: Access token lifetime in seconds (900 = 15 minutes)
- `user`: User information

**Error Responses:**

400 Bad Request - Invalid input:
```json
{
  "error": "invalid request payload"
}
```

401 Unauthorized - Invalid credentials:
```json
{
  "error": "invalid credentials"
}
```

403 Forbidden - Account locked:
```json
{
  "error": "account is locked"
}
```

403 Forbidden - Account inactive:
```json
{
  "error": "account is inactive"
}
```

**Security Features:**
- Failed login attempts are tracked
- Account is locked for 15 minutes after 5 failed attempts
- IP address and user agent are logged in audit trail

---

### 4. Refresh Token

Get a new access token using refresh token.

**Endpoint:** `POST /api/v1/auth/refresh`

**Authentication:** None (uses refresh token)

**Request Body:**
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "bmV3IHJlZnJlc2ggdG9rZW4...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Notes:**
- Old refresh token is automatically revoked
- New refresh token is issued (token rotation)
- New access token is generated

**Error Responses:**

400 Bad Request - Missing token:
```json
{
  "error": "invalid request payload"
}
```

401 Unauthorized - Invalid or expired token:
```json
{
  "error": "invalid or expired token"
}
```

---

### 5. Get Current User

Get authenticated user information.

**Endpoint:** `GET /api/v1/auth/me`

**Authentication:** Required (Bearer token)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "user",
  "is_verified": false,
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**

401 Unauthorized - Missing or invalid token:
```json
{
  "error": "missing authorization header"
}
```

404 Not Found - User not found:
```json
{
  "error": "user not found"
}
```

---

### 6. Logout

Logout from current session.

**Endpoint:** `POST /api/v1/auth/logout`

**Authentication:** Required (Bearer token)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body (optional):**
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."
}
```

**Success Response (204 No Content):**
No response body

**Notes:**
- If refresh token is provided, it will be revoked
- Logout event is logged in audit trail
- Access token remains valid until expiration (stateless JWT)

**Error Responses:**

401 Unauthorized - Invalid token:
```json
{
  "error": "invalid or expired token"
}
```

---

### 7. Logout All Devices

Logout from all devices (revoke all refresh tokens).

**Endpoint:** `POST /api/v1/auth/logout-all`

**Authentication:** Required (Bearer token)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Success Response (204 No Content):**
No response body

**Notes:**
- All user's refresh tokens are revoked
- All existing sessions will not be able to refresh tokens
- Access tokens remain valid until expiration
- Useful when account is compromised

**Error Responses:**

401 Unauthorized - Invalid token:
```json
{
  "error": "invalid or expired token"
}
```

---

### 8. Change Password

Change user password.

**Endpoint:** `PUT /api/v1/auth/change-password`

**Authentication:** Required (Bearer token)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "old_password": "StrongP@ss123",
  "new_password": "NewStrongP@ss456"
}
```

**Success Response (200 OK):**
```json
{
  "message": "password changed successfully"
}
```

**Notes:**
- Old password must be correct
- New password must meet strength requirements
- All refresh tokens are automatically revoked
- User needs to login again after password change

**Error Responses:**

400 Bad Request - Invalid old password:
```json
{
  "error": "invalid password"
}
```

400 Bad Request - Weak new password:
```json
{
  "error": "password is too weak"
}
```

401 Unauthorized - Invalid token:
```json
{
  "error": "invalid or expired token"
}
```

---

## Error Handling

All errors follow the same format:

```json
{
  "error": "error message description"
}
```

### HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Request successful, no response body
- `400 Bad Request` - Invalid input or validation error
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Authenticated but not authorized
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `500 Internal Server Error` - Server error

---

## Example Workflows

### Complete Authentication Flow

1. **Register:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecureP@ss123"
  }'
```

2. **Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecureP@ss123"
  }'
```

Save `access_token` and `refresh_token` from response.

3. **Access Protected Resource:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

4. **When Access Token Expires:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

5. **Logout:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### Password Change Flow

1. **Login first** (get access token)

2. **Change password:**
```bash
curl -X PUT http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "SecureP@ss123",
    "new_password": "NewSecureP@ss456"
  }'
```

3. **Login again with new password:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "NewSecureP@ss456"
  }'
```

---

## Security Considerations

### Token Management
- Access tokens expire in 15 minutes
- Refresh tokens expire in 30 days
- Store tokens securely (not in localStorage for web apps)
- Use HttpOnly cookies for refresh tokens when possible

### Password Security
- Passwords are hashed with bcrypt (cost 10)
- Never log or expose passwords
- Implement password strength requirements
- Consider password history to prevent reuse

### Rate Limiting
- Consider implementing rate limiting for login attempts
- Already has account lockout after 5 failed attempts

### CORS
- Configure `ALLOWED_ORIGINS` properly
- Don't use `*` in production

### HTTPS
- Always use HTTPS in production
- Tokens are bearer tokens and can be intercepted

---

## Testing with cURL

### Save tokens to variables (Linux/Mac):
```bash
# Login and extract tokens
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test@123456"}')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.access_token')
REFRESH_TOKEN=$(echo $RESPONSE | jq -r '.refresh_token')

# Use tokens
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Testing with Python:
```python
import requests

BASE_URL = "http://localhost:8080"

# Register
response = requests.post(f"{BASE_URL}/api/v1/auth/register", json={
    "email": "test@example.com",
    "password": "Test@123456"
})
print(response.json())

# Login
response = requests.post(f"{BASE_URL}/api/v1/auth/login", json={
    "email": "test@example.com",
    "password": "Test@123456"
})
data = response.json()
access_token = data['access_token']
refresh_token = data['refresh_token']

# Get user info
headers = {"Authorization": f"Bearer {access_token}"}
response = requests.get(f"{BASE_URL}/api/v1/auth/me", headers=headers)
print(response.json())
```

---

## Postman Collection

Import these as environment variables:
- `base_url`: `http://localhost:8080`
- `access_token`: (auto-filled from login)
- `refresh_token`: (auto-filled from login)

Use this script in Tests tab of Login request:
```javascript
if (pm.response.code === 200) {
    const response = pm.response.json();
    pm.environment.set("access_token", response.access_token);
    pm.environment.set("refresh_token", response.refresh_token);
}
```

Then use `{{access_token}}` in Authorization headers.
