package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func Info(c echo.Context) error {

	return c.JSON(http.StatusOK, "info")
}
