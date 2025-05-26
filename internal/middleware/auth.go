package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if the request has a valid token
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token")
		}

		// Validate the token (this is just a placeholder, implement your own logic)
		if !strings.HasPrefix(token, "Bearer ") {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		vals := strings.Split(token, " ")
		// expects token in the format "Bearer <username> <provider>"
		if len(vals) != 3 {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
		}

		username := vals[1]
		provider := vals[2]

		if username == "" || provider == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
		}

		c.Set("username", username)
		c.Set("provider", provider)

		return next(c)
	}
}
