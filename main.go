package main

import (
	"flag"
	"strings"

	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/zaker/anachrome-be/config"
	"github.com/zaker/anachrome-be/controllers"
	"github.com/zaker/anachrome-be/middleware"
	"github.com/zaker/anachrome-be/spa"


	"github.com/labstack/echo/v4"
	ec_middleware "github.com/labstack/echo/v4/middleware"
	"github.com/rjeczalik/notify"
)

var confFile = flag.String("c", "", "Path to config file")

func main() {
	flag.Parse()
	log.Println("Reading config from ", *confFile)
	conf, err := config.Load(*confFile)
	if err != nil {
		log.Fatal(err)
	}
	s := spa.New(conf.AppDir)
	absPath, err := filepath.Abs(conf.AppDir)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(absPath[:len(absPath)-5]+"/...", c, notify.Create|notify.Write); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)
	go func() {
		for {
			select {
			case ei := <-c:
				dirPath, fileName := filepath.Split(ei.Path())
				basePath := filepath.Base(dirPath)

				if basePath == "dist" && fileName == "index.html" {
					log.Println("Hit")
					go s.IndexParse()
				}
				//
			}
		}

	}()

	s.IndexParse()
	e := echo.New()
	e.Pre(ec_middleware.RemoveTrailingSlash())

	log.Println("Config:", conf)
	conf.GenerateCert(e)
	e.Use(ec_middleware.BodyLimit("2M"))
	if !conf.IsDebug {

		e.Use(ec_middleware.CSRF())
	}

	e.Use(ec_middleware.CORS())
	// e.Use(ec_middleware.CORSWithConfig(ec_middleware.CORSConfig{
	// 	// Skipper:      conf.DebugSkipper,
	// 	AllowOrigins: []string{conf.HostURI()},
	// 	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	// }))
	e.Use(ec_middleware.HTTPSRedirectWithConfig(ec_middleware.RedirectConfig{
		Skipper: func(c echo.Context) bool {

			if conf.IsDebug || strings.HasPrefix(c.Request().URL.Path, "/.well-known/") {
				return true
			}
			return false
		},
	}))

	e.Use(ec_middleware.Secure())
	e.Use(ec_middleware.GzipWithConfig(ec_middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(ec_middleware.Recover())
	e.Use(ec_middleware.Logger())
	e.Use(middleware.MIME())
	if !conf.IsDebug {

		e.Use(middleware.CSP())
		e.Use(middleware.HSTS())
	}
	// home route
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

	// Info
	e.Any("/info", controllers.Info)
	// GQL
	
	gql,err  := controllers.Init(conf)
	if err != nil{
		log.Fatal(err)
	}
	e.Any("/gql", echo.WrapHandler(gql.Handler()))

	if conf.HostName == "localhost" {
		e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(conf.HTTPSPort), ".tmp/cert.pem", ".tmp/key.pem"))
	} else if conf.Cert != "" {
		e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(conf.HTTPSPort), conf.Cert, conf.CertKey))
	} else {
		err = ConfigManager(&e.AutoTLSManager)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{
			Handler: e.AutoTLSManager.HTTPHandler(nil),
			Addr:    strconv.Itoa(conf.HTTPPort),
		}

		go s.ListenAndServe()
		e.Logger.Fatal(e.StartAutoTLS(":" + strconv.Itoa(conf.HTTPSPort)))
	}
}
