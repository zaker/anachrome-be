package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"./cert"
	"./middleware"
	"./spa"
	"github.com/labstack/echo"
	ec_middleware "github.com/labstack/echo/middleware"
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

//HostURI returns normalized uri for host
func (w *WebConfig) HostURI() string {
	uri := "https://" + w.HostName
	if w.HTTPSPort != 443 {
		uri += ":" + strconv.Itoa(w.HTTPSPort)
	}
	return uri
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
	e.Pre(ec_middleware.RemoveTrailingSlash())

	go func(e *echo.Echo) {
		e.GET("/", func(c echo.Context) (err error) {
			return c.Redirect(http.StatusMovedPermanently, conf.HostURI())
		})

		e.Logger.Fatal(e.Start(":" + strconv.Itoa(conf.HTTPPort)))
	}(e)

	log.Println("Config:", conf)
	if conf.HostName == "localhost" {
		if !fileExist(".tmp/cert.pem") || !fileExist(".tmp/key.pem") {
			cert.GenerateCertFiles("localhost", 365*24*time.Hour, true)
		}
	} else {

		e.Pre(ec_middleware.HTTPSRedirectWithConfig(ec_middleware.RedirectConfig{
			Skipper: func(c echo.Context) bool {

				if strings.HasPrefix(c.Request().URL.Path, "/.well-known/") {
					return true
				}
				return false
			},
		}))
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(conf.HostName)
		// Cache certificates
		e.AutoTLSManager.Prompt = autocert.AcceptTOS
		e.AutoTLSManager.Cache = autocert.DirCache(".cache")
	}
	e.Use(ec_middleware.BodyLimit("2M"))
	e.Use(ec_middleware.CSRF())
	e.Use(ec_middleware.CORSWithConfig(ec_middleware.CORSConfig{
		AllowOrigins: []string{conf.HostURI()},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(ec_middleware.Secure())
	e.Use(ec_middleware.GzipWithConfig(ec_middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(ec_middleware.Recover())
	e.Use(ec_middleware.Logger())
	e.Use(middleware.MIME())
	e.Use(middleware.CSP())
	e.Use(middleware.HSTS())
	e.Static("/", conf.AppDir)
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
		e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(conf.HTTPSPort), ".tmp/cert.pem", ".tmp/key.pem"))
	} else {

		e.Logger.Fatal(e.StartAutoTLS(":" + strconv.Itoa(conf.HTTPSPort)))
	}
}
