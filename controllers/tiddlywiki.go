package controllers

import (
	"bytes"
	"net/http"

	"github.com/zaker/anachrome-be/stores"

	"github.com/labstack/echo/v4"
)

type TW struct {
	Store *stores.TiddlerFileStore
}

func (tw *TW) Index(c echo.Context) error {

	switch c.Request().Method {
	case http.MethodGet:
		b := &bytes.Buffer{}
		err := tw.Store.Index(b)
		if err != nil {
			return err
		}
		return c.Stream(200, "text/html", b)
	case http.MethodHead:
		return c.Blob(200, "text/html", []byte{})
	default:

		return c.String(405, "Bad method")
	}

}

func (tw *TW) Status(c echo.Context) error {
	err := tw.Store.Status(c.Response())
	if err != nil {
		return err
	}

	return nil
}

func (tw *TW) List(c echo.Context) error {
	b := &bytes.Buffer{}
	err := tw.Store.List(b)
	if err != nil {
		return err
	}

	return c.Stream(200, "application/json", b)
}

func (tw *TW) Tiddler(c echo.Context) error {
	id := c.Param("id")
	switch c.Request().Method {
	case http.MethodGet:
		b := &bytes.Buffer{}
		err := tw.Store.Load(b, id)
		if err != nil {
			return err
		}

		return c.Stream(200, "application/json", b)
	case http.MethodPut:
		etag, err := tw.Store.Store(c.Request().Body, id)
		if err != nil {
			return err
		}

		c.Response().Header().Set("Etag", etag)

		return c.String(200, "Ok")
	default:

		return c.String(405, "Bad method")
	}
}

func (tw *TW) Delete(c echo.Context) error {

	err := tw.Store.Delete()
	if err != nil {
		return err
	}

	return c.String(200, "Ok")
}
