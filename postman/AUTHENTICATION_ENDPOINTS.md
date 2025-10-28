# Authentication Endpoints - Postman Guide

## 📋 Overview

The K8s Monitoring App now includes OAuth 2.0 authentication endpoints using Google. This guide explains how to test these endpoints in Postman.

## 🔐 Authentication Endpoints

### 1. Login Page
**Endpoint:** `GET /auth/login`

**Description:** Displays the login page with "Sign in with Google" button.

**Usage in Postman:**
```
GET {{base_url}}/auth/login
```

**Response:** HTML page with login UI

**Notes:**
- This is a web page, not an API endpoint
- Best viewed in a browser
- Shows the Google sign-in button

---

### 2. Initiate OAuth (Sign in with Google)
**Endpoint:** `GET /auth/google`

**Description:** Initiates the OAuth 2.0 flow with Google. Redirects to Google's authentication page.

**Usage in Postman:**
```
GET {{base_url}}/auth/google
```

**What happens:**
1. Generates a secure state token
2. Stores state in a cookie
3. Redirects to Google OAuth consent screen

**Response:** 307 Redirect to Google OAuth

**Notes:**
- This endpoint requires a browser for the full OAuth flow
- Postman will show the redirect URL but won't complete authentication
- To test fully, open this URL in a browser

**Testing in Browser:**
```bash
# Open in your default browser
open http://localhost:8080/auth/google

# Or with curl (shows redirect)
curl -v http://localhost:8080/auth/google
```

---

### 3. OAuth Callback
**Endpoint:** `GET /auth/callback`

**Query Parameters:**
- `code` (required) - Authorization code from Google OAuth
- `state` (required) - State token for CSRF protection

**Description:** OAuth 2.0 callback endpoint. Google redirects here after user authentication.

**Usage in Postman:**
```
GET {{base_url}}/auth/callback?code=AUTHORIZATION_CODE&state=STATE_TOKEN
```

**What happens:**
1. Validates the state token
2. Exchanges the authorization code for access token
3. Retrieves user information from Google
4. Validates email domain against `ALLOWED_GOOGLE_DOMAINS`
5. Creates a session in the database
6. Sets session cookie (`session_id`)
7. Redirects to dashboard

**Success Response:** 307 Redirect to `/` (dashboard)

**Error Response:** Redirect to `/auth/error?reason=ERROR_CODE`

**Notes:**
- This is automatically called by Google OAuth flow
- Manual testing is not recommended
- The authorization code is single-use and expires quickly

---

### 4. Logout
**Endpoint:** `GET /auth/logout`

**Description:** Logout endpoint. Deletes the user's session from the database and clears the session cookie.

**Usage in Postman:**
```
GET {{base_url}}/auth/logout
```

**What happens:**
1. Extracts session ID from cookie
2. Deletes session from database
3. Clears the `session_id` cookie
4. Redirects to `/auth/login`

**Response:** 307 Redirect to `/auth/login`

**Testing:**
```bash
# In browser (with active session)
open http://localhost:8080/auth/logout
```

---

### 5. Authentication Error Page
**Endpoint:** `GET /auth/error`

**Query Parameters:**
- `reason` (optional) - Error reason code

**Description:** Display authentication error page with specific error message.

**Usage in Postman:**
```
GET {{base_url}}/auth/error?reason=domain_not_allowed
```

**Possible Error Reasons:**

| Reason | Description |
|--------|-------------|
| `invalid_state` | Invalid authentication state (CSRF protection) |
| `no_code` | No authorization code received from Google |
| `exchange_failed` | Failed to exchange authorization code for token |
| `user_info_failed` | Failed to retrieve user information from Google |
| `email_not_verified` | Email address is not verified with Google |
| `domain_not_allowed` | Email domain is not in ALLOWED_GOOGLE_DOMAINS |
| `session_failed` | Failed to create session in database |
| `unauthorized` | Not authorized to access the application |

**Response:** HTML error page with user-friendly message

**Testing Different Errors:**
```
GET {{base_url}}/auth/error?reason=invalid_state
GET {{base_url}}/auth/error?reason=domain_not_allowed
GET {{base_url}}/auth/error?reason=email_not_verified
```

---

### 6. Check Session (Protected Route)
**Endpoint:** `GET /`

**Description:** Test if you have a valid session by accessing the dashboard (protected route).

**Usage in Postman:**
```
GET {{base_url}}/
```

**With Valid Session:**
- Status: 200 OK
- Response: Dashboard HTML

**Without Session:**
- Status: 307 Redirect
- Location: `/auth/login`

**Notes:**
- Requires a valid `session_id` cookie
- Sessions expire after 24 hours
- Session is extended on each request

---

## 🧪 Testing Authentication in Postman

### Method 1: Using Postman Interceptor (Recommended)

1. **Install Postman Interceptor**
   - Install the [Postman Interceptor](https://chrome.google.com/webstore/detail/postman-interceptor/) Chrome extension

2. **Enable Interceptor in Postman**
   - Click the satellite icon in Postman
   - Enable "Interceptor"
   - Enable "Sync cookies"

3. **Authenticate in Browser**
   - Open `http://localhost:8080` in Chrome
   - Complete the Google OAuth flow
   - You're now authenticated with a session cookie

4. **Test in Postman**
   - Postman will now use the same cookies as Chrome
   - You can access protected routes

### Method 2: Manual Cookie Management

1. **Get Session Cookie from Browser**
   - Authenticate in browser
   - Open Developer Tools (F12)
   - Go to Application > Cookies
   - Copy the `session_id` cookie value

2. **Add Cookie to Postman Request**
   - In Postman, go to the request
   - Click on "Cookies" (below Send button)
   - Add a new cookie:
     - Domain: `localhost`
     - Path: `/`
     - Name: `session_id`
     - Value: `[paste session ID]`

3. **Test Protected Routes**
   - Send requests to protected routes
   - They should now work with the session cookie

### Method 3: Browser-Based Testing (Simplest)

For OAuth flows, it's often easier to test in a browser:

```bash
# 1. Start the application
go run cmd/main.go

# 2. Test login flow
open http://localhost:8080

# 3. Test logout
open http://localhost:8080/auth/logout

# 4. Test error page
open http://localhost:8080/auth/error?reason=domain_not_allowed
```

---

## 🔍 Testing Scenarios

### Scenario 1: Successful Authentication Flow

1. **Access Dashboard (No Session)**
   ```
   GET {{base_url}}/
   → Redirects to /auth/login
   ```

2. **View Login Page**
   ```
   GET {{base_url}}/auth/login
   → Shows login page with Google button
   ```

3. **Initiate OAuth (In Browser)**
   ```
   Open: http://localhost:8080/auth/google
   → Redirects to Google
   → User authenticates
   → Redirects back to /auth/callback with code
   → Creates session
   → Redirects to dashboard
   ```

4. **Access Dashboard (With Session)**
   ```
   GET {{base_url}}/
   → Shows dashboard (200 OK)
   ```

5. **Logout**
   ```
   GET {{base_url}}/auth/logout
   → Clears session
   → Redirects to /auth/login
   ```

### Scenario 2: Domain Not Allowed

1. **User from non-allowed domain tries to login**
   - User authenticates with Google
   - Email domain is not in `ALLOWED_GOOGLE_DOMAINS`
   - Redirected to: `/auth/error?reason=domain_not_allowed`

### Scenario 3: Expired Session

1. **Session expires (after 24 hours)**
   ```
   GET {{base_url}}/api/v1/projects
   → Session expired
   → Redirects to /auth/login
   ```

### Scenario 4: API Request Without Authentication

1. **Try to access API endpoint without session**
   ```
   GET {{base_url}}/api/v1/projects
   → Returns 401 Unauthorized (JSON response for API)
   ```

---

## 🛠️ Configuration for Testing

### Environment Variables Required

```bash
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
ALLOWED_GOOGLE_DOMAINS=yourcompany.com

# Database
DB_CONNECTION_STRING=postgres://user:pass@localhost:5432/k8s_monitoring
```

### Postman Environment Variables

```json
{
  "base_url": "http://localhost:8080"
}
```

---

## 📊 Session Information

### Session Cookie Details

| Property | Value |
|----------|-------|
| Name | `session_id` |
| HttpOnly | `true` (cannot be accessed via JavaScript) |
| Secure | `true` (in production, HTTPS only) |
| SameSite | `Lax` (CSRF protection) |
| Max-Age | 24 hours |
| Path | `/` |

### Session Data in Database

```sql
-- View active sessions
SELECT * FROM sessions WHERE expires_at > NOW();

-- View specific user's session
SELECT * FROM sessions WHERE user_email = 'user@company.com';

-- Count active sessions
SELECT COUNT(*) FROM sessions WHERE expires_at > NOW();
```

---

## 🚨 Troubleshooting

### Issue: "OAuth not configured"

**Cause:** Missing OAuth environment variables

**Solution:**
```bash
export GOOGLE_CLIENT_ID="your-client-id"
export GOOGLE_CLIENT_SECRET="your-secret"
export GOOGLE_REDIRECT_URL="http://localhost:8080/auth/callback"
```

### Issue: "Domain not allowed"

**Cause:** User's email domain is not in `ALLOWED_GOOGLE_DOMAINS`

**Solution:**
```bash
export ALLOWED_GOOGLE_DOMAINS="company.com,partner.com"
```

### Issue: "Redirect URI mismatch"

**Cause:** `GOOGLE_REDIRECT_URL` doesn't match Google Cloud Console configuration

**Solution:**
1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Edit your OAuth 2.0 Client ID
3. Add the exact redirect URI: `http://localhost:8080/auth/callback`

### Issue: Cookies not working in Postman

**Solution:**
- Use Postman Interceptor (see Method 1 above)
- Or manually copy cookies from browser (see Method 2 above)
- Or test in browser (see Method 3 above)

### Issue: Session expires too quickly

**Cause:** Sessions expire after 24 hours of creation

**Solution:**
- Sessions are automatically extended on each request
- For testing, you can modify the expiry time in `internal/auth/session.go`

---

## 📚 Related Documentation

- **[OAuth Setup Guide](../docs/OAUTH_SETUP.md)** - Complete OAuth configuration
- **[OAuth Setup (PT)](../docs/OAUTH_SETUP_PT.md)** - Guia em português
- **[Authentication Summary](../docs/AUTHENTICATION_SUMMARY.md)** - Technical details
- **[Main README](../README.md)** - General documentation

---

## 💡 Tips for Testing

1. **Use Browser for OAuth Flow**
   - OAuth is designed for browsers
   - Redirects and cookies work automatically
   - Easier to debug visual issues

2. **Use Postman for API Testing**
   - After authenticating in browser
   - Use Interceptor to sync cookies
   - Test protected API endpoints

3. **Check Application Logs**
   - Logs show authentication attempts
   - Domain validation results
   - Session creation/deletion
   - Error details

4. **Monitor Database**
   - Check sessions table for active sessions
   - Verify session creation/deletion
   - Debug session expiry issues

5. **Test Different Scenarios**
   - Valid domain user
   - Invalid domain user
   - Unverified email
   - Expired session
   - No session

---

## ✅ Quick Test Checklist

- [ ] Health check works (no auth required)
- [ ] Dashboard redirects to login (no session)
- [ ] Login page loads correctly
- [ ] OAuth flow completes successfully
- [ ] Session cookie is set
- [ ] Dashboard loads (with session)
- [ ] API endpoints work (with session)
- [ ] Logout clears session
- [ ] Domain restriction works
- [ ] Error pages display correctly

---

**Note:** For the complete authentication flow to work, you must have valid Google OAuth credentials configured. See [docs/OAUTH_SETUP.md](../docs/OAUTH_SETUP.md) for setup instructions.

