package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware validates that the user has a valid session
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Skip auth for auth-related routes
			path := c.Path()
			if isPublicRoute(path) {
				return next(c)
			}

			// Get session cookie
			cookie, err := c.Cookie("session_id")
			if err != nil {
				log.Warn().Msg("No session cookie found")
				return redirectToLogin(c)
			}

			// Validate session
			session, err := GetSession(ctx, cookie.Value)
			if err != nil {
				log.Warn().Err(err).Str("session_id", cookie.Value).Msg("Invalid session")
				return redirectToLogin(c)
			}

			// Store session in context for use by handlers
			c.Set("session", session)
			c.Set("user_email", session.UserEmail)
			c.Set("user_name", session.UserName)
			c.Set("user_picture", session.UserPicture)

			// Extend session expiry on each request
			if err := UpdateSessionExpiry(ctx, session.ID); err != nil {
				log.Error().Str("session_id", session.ID).Msg("Failed to update session expiry")
			}

			return next(c)
		}
	}
}

// isPublicRoute checks if a route should bypass authentication
func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/auth/login",
		"/auth/google",
		"/auth/callback",
		"/auth/logout",
		"/auth/error",
		"/health",
	}

	for _, route := range publicRoutes {
		if path == route {
			return true
		}
	}

	// Allow static files
	if len(path) >= 7 && path[:7] == "/static" {
		return true
	}

	return false
}

// redirectToLogin redirects the user to the login page
func redirectToLogin(c echo.Context) error {
	// If it's an API request, return JSON error
	if isAPIRequest(c) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "Authentication required",
		})
	}

	// For web requests, redirect to login
	return c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
}

// isAPIRequest checks if the request is an API request
func isAPIRequest(c echo.Context) bool {
	path := c.Path()
	return len(path) >= 4 && path[:4] == "/api"
}

// GetUserFromContext retrieves the authenticated user from the context
func GetUserFromContext(c echo.Context) (email string, name string, picture string, ok bool) {
	emailVal := c.Get("user_email")
	nameVal := c.Get("user_name")
	pictureVal := c.Get("user_picture")

	if emailVal == nil || nameVal == nil {
		return "", "", "", false
	}

	email, emailOk := emailVal.(string)
	name, nameOk := nameVal.(string)

	// Picture is optional
	picture = ""
	if pictureVal != nil {
		picture, _ = pictureVal.(string)
	}

	return email, name, picture, emailOk && nameOk
}
