package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"./cert"
	"./spa"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
)

//WebConfig contains settings for webserver
type WebConfig struct {
	//HostName : hostname of  server
	HostName string
	//HTTPPort redirects from here
	HTTPPort int
	//HTTPSPort to here
	HTTPSPort int
	//AppDir path to SPA
	AppDir string
}

func initWebConfig(fileName string) WebConfig {
	if len(fileName) < 0 {
		fileName = "webConf.json"
	}

	conf := WebConfig{}
	err := func(fileName string) error {
		content, err := ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
		return json.Unmarshal(content, &conf)
	}(fileName)
	if err != nil {
		log.Println("Error reading config file", err)
		conf.HostName = "localhost"
		conf.HTTPPort = 8080
		conf.HTTPSPort = 8443
		m, err := json.Marshal(conf)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("webConf.json", m, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	return conf
}
func fileExist(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

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

var confFile = flag.String("c", "", "Path to config file")
var appDir = flag.String("a", "", "Path to App dist")
var bDir = flag.String("b", "", "Path to Blog")

func main() {
	flag.Parse()
	log.Println("Reading config from ", *confFile)
	conf := initWebConfig(*confFile)
	s := spa.New(conf.AppDir)
	s.IndexParse()
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	log.Println("Config:", conf)
	if conf.HostName == "localhost" {
		if !fileExist("cert.pem") || !fileExist("key.pem") {
			cert.GenerateCertFiles("localhost", 365*24*time.Hour, false)
		}
	} else {
		e.Pre(middleware.HTTPSRedirect())
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("anachro.me")
		// Cache certificates
		e.AutoTLSManager.Prompt = autocert.AcceptTOS
		e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	}
	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.CSRF())
	e.Use(middleware.Secure())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(MIME())
	e.Use(CSP())
	e.Static("/", "../anachrome-fe/dist")
	e.GET("/", func(c echo.Context) (err error) {
		pusher, ok := c.Response().Writer.(http.Pusher)
		if ok {
			for _, f := range s.PushFiles {
				if err = pusher.Push(f, nil); err != nil {
					return
				}
			}
		}
		return c.File(s.IndexPath)
	})

	if conf.HostName == "localhost" {
		e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(conf.HTTPSPort), "cert.pem", "key.pem"))
	} else {
		go func(c *echo.Echo) {
			e.Logger.Fatal(e.Start(":" + strconv.Itoa(conf.HTTPPort)))
		}(e)
		e.Logger.Fatal(e.StartAutoTLS(":" + strconv.Itoa(conf.HTTPSPort)))
	}
}
