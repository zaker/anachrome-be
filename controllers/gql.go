package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func GQL(c echo.Context) error {

	return c.String(http.StatusOK, "gqlolol")
}
