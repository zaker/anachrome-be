package servers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zaker/anachrome-be/controllers"
	"github.com/zaker/anachrome-be/middleware"
	"github.com/zaker/anachrome-be/services"

	"github.com/labstack/echo/v4"
	"github.com/rjeczalik/notify"

	// jwt "github.com/dgrijalva/jwt-go"
	ec_middleware "github.com/labstack/echo/v4/middleware"
)

type serverMode int

func (sm serverMode) String() string {
	switch sm {
	case NONE:
		return "None"
	case INSECURE:
		return "Insecure"
	case SELFSIGNED:
		return "With selfsigned"
	case SECURE:
		return "Lets Encrypt"
	case LETSENCRYPT:
		return "Secure"
	default:
		return "Unknown"
	}

}

const (
	NONE serverMode = iota
	INSECURE
	SELFSIGNED
	SECURE
	LETSENCRYPT
)

type APIServer struct {
	serv     APIService
	app      *echo.Echo
	wc       WebConfig
	version  string
	hostAddr string
	devMode  bool
}

type APIService struct {
	spa *services.SPA
	gql *services.GQL
}
type WebConfig struct {
	Mode       serverMode
	HostName   string
	HttpPort   int
	HttpsPort  int
	Cert       string
	CertKey    string
	domains    string
	domainmail string
}

type Option interface {
	apply(*APIServer) error
}

func DefaultAPIServer() *APIServer {

	app := echo.New()
	return &APIServer{
		app:      app,
		hostAddr: "localhost:8080"}
}

func NewHTTPServer(opts ...Option) (hs *APIServer, err error) {
	hs = DefaultAPIServer()
	for _, opt := range opts {
		err = opt.apply(hs)
		if err != nil {
			return nil, fmt.Errorf("Applying config failed: %w", err)
		}
	}
	hs.app.Pre(ec_middleware.RemoveTrailingSlash())

	services.GenerateCert(hs.hostAddr)

	hs.app.Use(ec_middleware.BodyLimit("2M"))
	if !hs.devMode {

		hs.app.Use(ec_middleware.CSRF())
	}

	hs.app.Use(ec_middleware.CORS())

	hs.app.Use(ec_middleware.Secure())
	hs.app.Use(ec_middleware.GzipWithConfig(ec_middleware.GzipConfig{
		Level: 5,
	}))
	hs.app.Use(ec_middleware.Recover())
	hs.app.Use(ec_middleware.Logger())
	hs.app.Use(middleware.MIME())
	if !hs.devMode {

		hs.app.Use(middleware.CSP())
		hs.app.Use(middleware.HSTS())
	}

	return hs, nil
}

func (as *APIServer) registerEndpoints() {

	// SPA
	if as.serv.spa != nil {
		s := as.serv.spa
		as.app.Static("/", s.AppDir())
		as.app.GET("/", func(c echo.Context) (err error) {
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
	}

	// GQL
	if as.serv.gql != nil {
		handler := controllers.GQLHandler(as.serv.gql)
		as.app.Any("/gql", echo.WrapHandler(handler()))
	}

}

func (as *APIServer) Serve() error {

	as.registerEndpoints()

	wc := as.wc
	fmt.Println("webconf", wc)
	switch wc.Mode {
	case INSECURE:
		return as.app.Start("localhost:" + strconv.Itoa(wc.HttpPort))
	case SELFSIGNED:
		return as.app.StartTLS(":"+strconv.Itoa(wc.HttpsPort), wc.Cert, wc.CertKey)
	case LETSENCRYPT:
		return as.app.StartAutoTLS(":" + strconv.Itoa(wc.HttpsPort))
	case SECURE:
		return as.app.StartTLS(":"+strconv.Itoa(wc.HttpsPort), wc.Cert, wc.CertKey)
	default:
		return fmt.Errorf("no http server mode chosen")
	}
}

func modeSelect(wc WebConfig) serverMode {
	m := INSECURE
	if strings.HasPrefix(wc.HostName, "localhost") {

		if wc.HttpsPort > 0 {
			m = SELFSIGNED
		}

	} else {
		m = SECURE
	}
	return m
}

func WithWebConfig(wc WebConfig) Option {

	return newFuncOption(func(as *APIServer) (err error) {
		if wc.Mode == NONE {
			wc.Mode = modeSelect(wc)
		}
		as.wc = wc

		return
	})
}

func WithAPIVersion(version string) Option {

	return newFuncOption(func(hs *APIServer) (err error) {
		hs.version = version
		return
	})
}

func WithHTTPOnly() Option {

	return newFuncOption(func(hs *APIServer) (err error) {
		hs.wc.Mode = INSECURE
		return
	})
}

func WithSPA(appDir string) Option {

	return newFuncOption(func(hs *APIServer) (err error) {
		hs.serv.spa, err = services.NewSPA(appDir)
		if err != nil {
			return err
		}
		absPath, err := filepath.Abs(appDir)
		if err != nil {
			return err
		}
		c := make(chan notify.EventInfo, 1)
		if err := notify.Watch(absPath[:len(absPath)-5]+"/...", c, notify.Create|notify.Write); err != nil {
			return err
		}
		defer notify.Stop(c)
		go func() {
			for ei := range c {
				dirPath, fileName := filepath.Split(ei.Path())
				basePath := filepath.Base(dirPath)

				if basePath == "dist" && fileName == "index.html" {
					log.Println("Hit")
					go hs.serv.spa.IndexParse()
				}
				//
			}

		}()

		hs.serv.spa.IndexParse()
		return
	})
}

func WithDevMode() Option {

	return newFuncOption(func(hs *APIServer) (err error) {
		hs.devMode = true
		return
	})
}

func WithGQL(devMode bool) Option {

	return newFuncOption(func(as *APIServer) (err error) {

		gql, err := services.InitGQL(devMode)
		if err != nil {
			return
		}
		as.serv.gql = gql
		return
	})
}

// func WithTLS(certFile, keyFile string) Option {

// 	return newFuncOption(func(hs *APIServer) (err error) {

// 		if len(certFile) == 0 {
// 			return fmt.Errorf("No cert file selected for TLS")
// 		}

// 		if len(keyFile) == 0 {
// 			return fmt.Errorf("No key file selected for TLS")
// 		}
// 		hs.chosenMode = SECURE
// 		hs.certFile = certFile
// 		hs.privKeyFile = keyFile
// 		return
// 	})
// }
func WithLetsEncrypt(domains, domainmail string) Option {

	return newFuncOption(func(hs *APIServer) (err error) {
		if len(domains) == 0 {
			return fmt.Errorf("No domains selected for LetsEncrypt")
		}

		if len(domainmail) == 0 {
			return fmt.Errorf("No domain mail selected for LetsEncrypt")
		}
		hs.wc.Mode = LETSENCRYPT
		hs.wc.domains = domains
		hs.wc.domainmail = domainmail
		return
	})
}
