# OAuth 2.0 Authentication Setup

This application uses Google OAuth 2.0 for authentication. Only users from authorized Google domains can access the application.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Google Cloud Setup](#google-cloud-setup)
- [Environment Configuration](#environment-configuration)
- [Testing Authentication](#testing-authentication)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- A Google Cloud Platform account
- Admin access to your organization's Google Workspace (if restricting by domain)
- Access to the application's environment variables

## Google Cloud Setup

### 1. Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Note your project ID

### 2. Enable Google+ API

1. In the Google Cloud Console, go to **APIs & Services** > **Library**
2. Search for "Google+ API"
3. Click **Enable**

### 3. Configure OAuth Consent Screen

1. Go to **APIs & Services** > **OAuth consent screen**
2. Choose **Internal** if you want to restrict to your organization only, or **External** for public access
3. Fill in the required fields:
   - **App name**: K8s Monitoring App
   - **User support email**: Your email
   - **Developer contact information**: Your email
4. Click **Save and Continue**
5. On the Scopes page, click **Add or Remove Scopes**
6. Add the following scopes:
   - `userinfo.email`
   - `userinfo.profile`
7. Click **Save and Continue**
8. Review and click **Back to Dashboard**

### 4. Create OAuth 2.0 Credentials

1. Go to **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **OAuth client ID**
3. Choose **Web application**
4. Configure the following:
   - **Name**: K8s Monitoring App
   - **Authorized JavaScript origins**: 
     - `http://localhost:8080` (for local development)
     - `https://your-domain.com` (for production)
   - **Authorized redirect URIs**:
     - `http://localhost:8080/auth/callback` (for local development)
     - `https://your-domain.com/auth/callback` (for production)
5. Click **Create**
6. **Important**: Copy the **Client ID** and **Client Secret** - you'll need these for environment variables

## Environment Configuration

Add the following environment variables to your `.env` file or deployment configuration:

```bash
# OAuth Configuration
GOOGLE_CLIENT_ID=your-client-id-here.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret-here
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
ALLOWED_GOOGLE_DOMAINS=yourdomain.com,anotherdomain.com
ALLOWED_GOOGLE_EMAILS=user1@yourdomain.com,user2@anotherdomain.com
```

### Environment Variables Explained

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `GOOGLE_CLIENT_ID` | Yes | OAuth 2.0 Client ID from Google Cloud Console | `123456789-abc.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | Yes | OAuth 2.0 Client Secret from Google Cloud Console | `GOCSPX-abc123def456` |
| `GOOGLE_REDIRECT_URL` | Yes | Callback URL for OAuth flow (must match Google Console) | `https://your-domain.com/auth/callback` |
| `ALLOWED_GOOGLE_DOMAINS` | No | Comma-separated list of allowed email domains. If empty, all domains are allowed | `company.com,partner.com` |
| `ALLOWED_GOOGLE_EMAILS` | No | Comma-separated list of allowed email addresses. If set, only listed emails can login (in addition to domain rules) | `user1@company.com,user2@partner.com` |

### Production Configuration Example

```bash
# Production OAuth Configuration
GOOGLE_CLIENT_ID=123456789-abc.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-abc123def456
GOOGLE_REDIRECT_URL=https://monitoring.yourcompany.com/auth/callback
ALLOWED_GOOGLE_DOMAINS=yourcompany.com
ALLOWED_GOOGLE_EMAILS=admin@yourcompany.com
```

### Local Development Example

```bash
# Local Development OAuth Configuration
GOOGLE_CLIENT_ID=123456789-xyz.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xyz789uvw012
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
ALLOWED_GOOGLE_DOMAINS=yourcompany.com,gmail.com
ALLOWED_GOOGLE_EMAILS=
```

## Database Migration

The authentication system requires a sessions table. This will be created automatically when the application starts, but you can also run migrations manually:

```bash
# Migrations are applied automatically on startup
# The session table schema is in:
# database/migrations/1761700000_add_sessions_table.up.sql
```

### Sessions Table Schema

```sql
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expiry DATETIME,
    created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
    expires_at DATETIME NOT NULL
);
```

## Testing Authentication

### 1. Start the Application

```bash
# Make sure your environment variables are set
source .env

# Run the application
go run cmd/main.go
```

### 2. Access the Login Page

Navigate to: `http://localhost:8080/auth/login`

You should see the login page with a "Sign in with Google" button.

### 3. Test the Login Flow

1. Click "Sign in with Google"
2. You'll be redirected to Google's authentication page
3. Sign in with a Google account
4. Grant permissions to the application
5. You should be redirected back to the application dashboard

### 4. Verify Session

After successful login:
- A session cookie (`session_id`) should be set in your browser
- You can access protected routes (dashboard, metrics, etc.)
- The session is valid for 24 hours

### 5. Test Logout

Navigate to: `http://localhost:8080/auth/logout`

You should be redirected to the login page, and your session should be cleared.

## Authentication Flow

```
┌─────────┐                ┌──────────────┐              ┌────────┐
│ Browser │                │  Application │              │ Google │
└────┬────┘                └──────┬───────┘              └───┬────┘
     │                            │                          │
     │  1. Access protected page  │                          │
     ├───────────────────────────>│                          │
     │                            │                          │
     │  2. Redirect to login      │                          │
     │<───────────────────────────┤                          │
     │                            │                          │
     │  3. Click "Sign in"        │                          │
     ├───────────────────────────>│                          │
     │                            │                          │
     │  4. Redirect to Google     │                          │
     │<───────────────────────────┤                          │
     │                            │                          │
     │  5. Authenticate with Google                          │
     ├──────────────────────────────────────────────────────>│
     │                            │                          │
     │  6. Redirect with auth code                           │
     │<──────────────────────────────────────────────────────┤
     │                            │                          │
     │  7. Send code to callback  │                          │
     ├───────────────────────────>│                          │
     │                            │                          │
     │                            │  8. Exchange code        │
     │                            ├─────────────────────────>│
     │                            │                          │
     │                            │  9. Return tokens        │
     │                            │<─────────────────────────┤
     │                            │                          │
     │                            │ 10. Get user info        │
     │                            ├─────────────────────────>│
     │                            │                          │
     │                            │ 11. Return user data     │
     │                            │<─────────────────────────┤
     │                            │                          │
     │                            │ 12. Validate domain      │
     │                            │                          │
     │                            │ 13. Create session       │
     │                            │                          │
     │ 14. Set session cookie     │                          │
     │<───────────────────────────┤                          │
     │                            │                          │
     │ 15. Redirect to dashboard  │                          │
     │<───────────────────────────┤                          │
     │                            │                          │
```

## Protected Routes

All routes except the following are protected by authentication:

**Public Routes (no authentication required):**
- `/health` - Health check endpoint
- `/auth/login` - Login page
- `/auth/google` - OAuth initiation (redirects to Google)
- `/auth/callback` - OAuth callback (handles Google response)
- `/auth/logout` - Logout endpoint
- `/auth/error` - Authentication error page
- `/static/*` - Static assets (CSS, JS, images)

**Protected Routes (authentication required):**
- `/` - Dashboard
- `/api/v1/*` - All REST API endpoints
- `/api/ui/*` - UI API endpoints

## Security Features

### Session Management
- Sessions are stored in PostgreSQL
- Session IDs are cryptographically secure (32-byte random)
- Sessions expire after 24 hours
- Session expiry is extended on each request
- Logout clears the session from the database

### Cookie Security
- HttpOnly cookies (prevents JavaScript access)
- Secure flag in production (HTTPS only)
- SameSite=Lax (CSRF protection)

### Domain Validation
- Email domain is validated against `ALLOWED_GOOGLE_DOMAINS`
- Email must be verified by Google
- Hosted domain (`hd`) is checked for G Suite accounts

## Troubleshooting

### Error: "OAuth not configured"

**Cause**: Missing or invalid OAuth environment variables

**Solution**: 
1. Check that `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are set
2. Verify the values are correct (copy from Google Cloud Console)
3. Restart the application after setting environment variables

### Error: "Domain not allowed"

**Cause**: User's email domain is not in the `ALLOWED_GOOGLE_DOMAINS` list

**Solution**:
1. Add the user's domain to `ALLOWED_GOOGLE_DOMAINS`
2. Format: `ALLOWED_GOOGLE_DOMAINS=domain1.com,domain2.com`
3. Restart the application
4. Have the user try logging in again

### Error: "Redirect URI mismatch"

**Cause**: The redirect URI doesn't match what's configured in Google Cloud Console

**Solution**:
1. Check `GOOGLE_REDIRECT_URL` environment variable
2. Go to Google Cloud Console > APIs & Services > Credentials
3. Edit your OAuth 2.0 Client ID
4. Add the exact redirect URI to "Authorized redirect URIs"
5. Format: `http://localhost:8080/auth/callback` or `https://your-domain.com/auth/callback`

### Error: "Session not found or expired"

**Cause**: Session has expired or was deleted

**Solution**:
1. User should log in again
2. Sessions expire after 24 hours
3. Check database connectivity
4. Verify the sessions table exists

### Sessions are not persisting

**Cause**: Database connection issues or missing sessions table

**Solution**:
1. Check database connection string: `DB_CONNECTION_STRING`
2. Verify the application can connect to PostgreSQL
3. Check if the sessions table exists:
   ```sql
   SELECT * FROM information_schema.tables WHERE table_name = 'sessions';
   ```
4. Run migrations if the table is missing

### Users are not being redirected after login

**Cause**: Incorrect redirect URL configuration

**Solution**:
1. Check browser console for errors
2. Verify `GOOGLE_REDIRECT_URL` matches the callback route
3. Ensure the URL is accessible (not blocked by firewall)

## API Access with Authentication

If you need to access the REST API endpoints programmatically:

### Option 1: Session Cookie (for web applications)

After successful authentication, the `session_id` cookie is automatically included in requests.

### Option 2: Admin Token (for service accounts)

For automated scripts or services, you can still use the `ADMIN_TOKEN` method:

```bash
curl -H "Authorization: YOUR_ADMIN_TOKEN" \
  http://localhost:8080/api/v1/projects
```

**Note**: The admin token auth is separate from OAuth and should be used for service-to-service communication.

## Cleanup and Maintenance

### Session Cleanup

Sessions are automatically cleaned up when they expire. To manually clean up old sessions:

```sql
-- Delete expired sessions
DELETE FROM sessions WHERE expires_at <= NOW();
```

### Monitoring

Check active sessions:

```sql
-- Count active sessions
SELECT COUNT(*) FROM sessions WHERE expires_at > NOW();

-- List active sessions
SELECT user_email, created_at, expires_at 
FROM sessions 
WHERE expires_at > NOW()
ORDER BY created_at DESC;
```

## Additional Resources

- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [Google Cloud Console](https://console.cloud.google.com/)
- [OAuth 2.0 Playground](https://developers.google.com/oauthplayground/)

## Support

For issues or questions:
1. Check the application logs for detailed error messages
2. Verify all environment variables are correctly set
3. Ensure the Google Cloud Console configuration matches your setup
4. Contact your system administrator for access issues
