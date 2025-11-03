package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s-monitoring-app/internal/env"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	HD            string `json:"hd"` // Hosted domain - the domain of the user's email
}

type OAuthConfig struct {
	Config         *oauth2.Config
	AllowedDomains []string
	AllowedEmails  []string
}

var OAuth *OAuthConfig

// InitOAuth initializes the OAuth configuration
func InitOAuth() error {
	if env.GOOGLE_CLIENT_ID == "" || env.GOOGLE_CLIENT_SECRET == "" {
		return fmt.Errorf("google oauth credentials not configured")
	}

	OAuth = &OAuthConfig{
		Config: &oauth2.Config{
			ClientID:     env.GOOGLE_CLIENT_ID,
			ClientSecret: env.GOOGLE_CLIENT_SECRET,
			RedirectURL:  env.GOOGLE_REDIRECT_URL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		AllowedDomains: parseAllowedDomains(env.ALLOWED_GOOGLE_DOMAINS),
		AllowedEmails:  parseAllowedEmails(env.ALLOWED_GOOGLE_EMAILS),
	}

	log.Info().
		Str("redirect_url", env.GOOGLE_REDIRECT_URL).
		Strs("allowed_domains", OAuth.AllowedDomains).
		Strs("allowed_emails", OAuth.AllowedEmails).
		Msg("OAuth initialized successfully")

	return nil
}

// parseAllowedDomains parses comma-separated domains
func parseAllowedDomains(domains string) []string {
	if domains == "" {
		return []string{}
	}

	parts := strings.Split(domains, ",")
	result := make([]string, 0, len(parts))
	for _, domain := range parts {
		trimmed := strings.TrimSpace(domain)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// parseAllowedEmails parses comma-separated emails
func parseAllowedEmails(emails string) []string {
	if emails == "" {
		return []string{}
	}

	parts := strings.Split(emails, ",")
	result := make([]string, 0, len(parts))
	for _, email := range parts {
		trimmed := strings.TrimSpace(email)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GenerateStateToken generates a random state token for OAuth
func GenerateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL returns the OAuth authorization URL
func (o *OAuthConfig) GetAuthURL(state string) string {
	return o.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// ExchangeCode exchanges the authorization code for a token
func (o *OAuthConfig) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return o.Config.Exchange(ctx, code)
}

// GetUserInfo retrieves user information from Google
func (o *OAuthConfig) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := o.Config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// IsDomainAllowed checks if the email domain is in the allowed list
func (o *OAuthConfig) IsDomainAllowed(email string) bool {
	// If no domains configured, allow all (for backward compatibility)
	if len(o.AllowedDomains) == 0 {
		return true
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])
	for _, allowedDomain := range o.AllowedDomains {
		if strings.TrimSpace(strings.ToLower(allowedDomain)) == strings.TrimSpace(strings.ToLower(domain)) {
			return true
		}
	}

	log.Warn().Str("email", email).Strs("allowed_domains", o.AllowedDomains).Msg("Email domain not in allowed list")
	return false
}

// IsAddressAllowed checks if the full email address is in the allowed list
func (o *OAuthConfig) IsAddressAllowed(email string) bool {
	// If no emails configured, allow all (domain rule will still apply separately)
	if len(o.AllowedEmails) == 0 {
		return true
	}

	for _, allowed := range o.AllowedEmails {
		if strings.EqualFold(strings.TrimSpace(allowed), strings.TrimSpace(email)) {
			return true
		}
	}

	log.Warn().Str("email", email).Strs("allowed_emails", o.AllowedEmails).Msg("Email address not in allowed list")
	return false
}

// HandleLogin handles the OAuth login initiation
func HandleLogin(c echo.Context) error {
	if OAuth == nil {
		return c.String(http.StatusInternalServerError, "OAuth not configured")
	}

	state, err := GenerateStateToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate state token")
		return c.String(http.StatusInternalServerError, "Failed to generate state token")
	}

	// Store state in cookie for validation
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   env.ENV == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	}
	c.SetCookie(cookie)

	authURL := OAuth.GetAuthURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleCallback handles the OAuth callback
func HandleCallback(c echo.Context) error {
	if OAuth == nil {
		return c.String(http.StatusInternalServerError, "OAuth not configured")
	}

	ctx := c.Request().Context()

	// Validate state
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil {
		log.Error().Msg("State cookie not found")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=invalid_state")
	}

	state := c.QueryParam("state")
	if state == "" || state != stateCookie.Value {
		log.Error().Msg("Invalid state parameter")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=invalid_state")
	}

	// Clear state cookie
	c.SetCookie(&http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange code for token
	code := c.QueryParam("code")
	if code == "" {
		log.Error().Msg("Authorization code not found")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=no_code")
	}

	token, err := OAuth.ExchangeCode(ctx, code)
	if err != nil {
		log.Error().Msg("Failed to exchange code for token")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=exchange_failed")
	}

	// Get user info
	userInfo, err := OAuth.GetUserInfo(ctx, token)
	if err != nil {
		log.Error().Msg("Failed to get user info")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=user_info_failed")
	}

	// Verify email
	if !userInfo.EmailVerified {
		log.Warn().Str("email", userInfo.Email).Msg("Email not verified")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=email_not_verified")
	}

	// Check if email domain is allowed (if configured)
	if !OAuth.IsDomainAllowed(userInfo.Email) {
		log.Warn().
			Str("email", userInfo.Email).
			Str("domain", userInfo.HD).
			Msg("Domain not allowed")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=domain_not_allowed")
	}

	// Check if specific email address is allowed (if configured)
	if !OAuth.IsAddressAllowed(userInfo.Email) {
		log.Warn().
			Str("email", userInfo.Email).
			Msg("Email not allowed")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=email_not_allowed")
	}

	// Create session
	session, err := CreateSession(ctx, userInfo.Email, userInfo.Name, userInfo.Picture, token)
	if err != nil {
		log.Error().Msg("Failed to create session")
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/error?reason=session_failed")
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   env.ENV == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
	}
	c.SetCookie(cookie)

	log.Info().
		Str("email", userInfo.Email).
		Str("session_id", session.ID).
		Msg("User authenticated successfully")

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

// HandleLogout handles user logout
func HandleLogout(c echo.Context) error {
	ctx := c.Request().Context()

	// Get session from cookie
	cookie, err := c.Cookie("session_id")
	if err == nil {
		// Delete session from database
		if err := DeleteSession(ctx, cookie.Value); err != nil {
			log.Error().Str("session_id", cookie.Value).Msg("Failed to delete session")
		}
	}

	// Clear session cookie
	c.SetCookie(&http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	return c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
}

// HandleAuthError displays authentication error page
func HandleAuthError(c echo.Context) error {
	reason := c.QueryParam("reason")

	messages := map[string]string{
		"invalid_state":      "Invalid authentication state. Please try again.",
		"no_code":            "No authorization code received. Please try again.",
		"exchange_failed":    "Failed to exchange authorization code. Please try again.",
		"user_info_failed":   "Failed to retrieve user information. Please try again.",
		"email_not_verified": "Your email address is not verified with Google.",
		"domain_not_allowed": "Your email domain is not authorized to access this application.",
		"email_not_allowed":  "Your email address is not authorized to access this application.",
		"session_failed":     "Failed to create session. Please try again.",
		"unauthorized":       "You are not authorized to access this application.",
	}

	message, ok := messages[reason]
	if !ok {
		message = "An unknown error occurred during authentication."
	}

	return c.Render(http.StatusUnauthorized, "auth-error.html", map[string]interface{}{
		"Message": message,
		"Reason":  reason,
	})
}
