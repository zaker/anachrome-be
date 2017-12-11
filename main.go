package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	e := echo.New()
	e.Pre(middleware.HTTPSRedirect())
	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("anachro.me")
	// Cache certificates
	e.AutoTLSManager.Prompt = autocert.AcceptTOS
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `                                                               
                        <h1>Welcome to Echo!</h1>                                                            
                        <h3>TLS certificates automatically installed from Let's Encrypt :)</h3>              
                `)
	})
	go func(c *echo.Echo) {
		e.Logger.Fatal(e.Start(":80"))
	}(e)

	e.Logger.Fatal(e.StartAutoTLS(":443"))
}
