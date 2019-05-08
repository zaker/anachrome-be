package middleware

import (
	"path/filepath"

	"github.com/labstack/echo/v4"
)

var mimeTypes = map[string]string{
	".js":   "text/javascript",
	".css":  "text/css",
	".html": "text/html",
	".ico":  "image/x-icon",
	"":      "text/html"}

//MIME middleware sets content type headers based on extension
func MIME() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			p := filepath.Ext(c.Request().URL.Path)
			typ, ok := mimeTypes[p]
			if ok && len(typ) > 0 {
				c.Response().Header().Set(echo.HeaderContentType, typ)
			}
			return next(c)
		}

	}
}

//CSP middleware sets content security policy
func CSP() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			p := filepath.Ext(c.Request().URL.Path)
			typ, ok := mimeTypes[p]
			if ok && len(typ) > 0 {
				c.Response().Header().Set(
					echo.HeaderContentSecurityPolicy,
					"default-src 'self';img-src 'self' data:;style-src 'self' 'unsafe-inline'")
			}
			return next(c)
		}

	}
}

//HSTS middleware sets strict transport security
func HSTS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(
				echo.HeaderStrictTransportSecurity,
				"max-age=10886400; includeSubDomains")

			return next(c)
		}

	}
}
