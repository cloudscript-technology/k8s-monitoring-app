package server

import (
	"net/http"

	"k8s-monitoring-app/internal/env"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func AdminAuthMiddleware() echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || authHeader != env.ADMIN_TOKEN {
				log.Warn().Msg("authorizing error: authorization header not found")
				return c.String(http.StatusUnauthorized, "Not authorized")
			}

			return h(c)
		}
	}
}
