package servers

import (
	"fmt"
	"strconv"

	"github.com/zaker/anachrome-be/stores"

	"github.com/zaker/anachrome-be/controllers"
	"github.com/zaker/anachrome-be/middleware"
	"github.com/zaker/anachrome-be/services"

	"github.com/labstack/echo/v4"

	// jwt "github.com/dgrijalva/jwt-go"
	ec_middleware "github.com/labstack/echo/v4/middleware"
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
	gql   *services.GQL
	bs    stores.BlogStore
	authn *services.WebAuthN
}
type WebConfig struct {
	HostName string
	HttpPort int
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

	// GQL
	if as.serv.gql != nil {
		handler := controllers.GQLHandler(as.serv.gql)
		as.app.Any("/gql", echo.WrapHandler(handler()))
	}

	// WebAuthN
	if as.serv.authn != nil {
		authCtl := &controllers.Auth{
			Service: as.serv.authn,
		}
		cAuth := as.app.Group("/auth")
		cAuth.POST("/register-begin", authCtl.BeginRegistration)
		cAuth.POST("/register-finish", authCtl.FinishRegistration)
		cAuth.POST("/login-begin", authCtl.BeginLogin)
		cAuth.POST("/login-finish", authCtl.FinishLogin)
	}

}

func (as *APIServer) Serve() error {

	as.registerEndpoints()

	wc := as.wc
	fmt.Println("webconf", wc)

	return as.app.Start(":" + strconv.Itoa(wc.HttpPort))

}

func WithWebConfig(wc WebConfig) Option {

	return newFuncOption(func(as *APIServer) (err error) {

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

func WithDevMode() Option {

	return newFuncOption(func(hs *APIServer) (err error) {
		hs.devMode = true
		return
	})
}

func WithGQL(devMode bool, blogStore stores.BlogStore) Option {

	return newFuncOption(func(as *APIServer) (err error) {

		gql, err := services.InitGQL(devMode, blogStore)
		if err != nil {
			return
		}
		as.serv.gql = gql
		return
	})
}

func WithBlogStore(blogStore stores.BlogStore) Option {

	return newFuncOption(func(as *APIServer) (err error) {

		as.serv.bs = blogStore
		return
	})
}
func WithAuthN(authn *services.WebAuthN) Option {

	return newFuncOption(func(hs *APIServer) (err error) {
		if authn == nil {
			return fmt.Errorf("Cannot set auth service to nil")
		}

		hs.serv.authn = authn
		return
	})
}
