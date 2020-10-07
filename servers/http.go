package servers

import (
	"fmt"
	"strconv"

	"github.com/zaker/anachrome-be/stores/blog"

	"github.com/zaker/anachrome-be/controllers"
	"github.com/zaker/anachrome-be/middleware"
	"github.com/zaker/anachrome-be/services"

	"github.com/labstack/echo/v4"

	// jwt "github.com/dgrijalva/jwt-go"
	ec_middleware "github.com/labstack/echo/v4/middleware"
)

type APIServer struct {
	serv     Services
	app      *echo.Echo
	wc       WebConfig
	version  string
	hostAddr string
}

type Services struct {
	blogStore blog.BlogStore
	authn     *services.WebAuthN
}
type WebConfig struct {
	HostName  string
	HTTPPort  int
	enableGQL bool
	devMode   bool
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
	if !hs.wc.devMode {

		hs.app.Use(ec_middleware.CSRFWithConfig(ec_middleware.CSRFConfig{
			Skipper: func(ctx echo.Context) bool {
				if ctx.Path() == "/gql" {
					return true
				}
				return false
			},
			TokenLookup:    "header:X-XSRF-TOKEN",
			CookieSecure:   false,
			CookieHTTPOnly: false,
		}))
	}

	hs.app.Use(ec_middleware.CORS())

	hs.app.Use(ec_middleware.Secure())
	hs.app.Use(ec_middleware.GzipWithConfig(ec_middleware.GzipConfig{
		Level: 5,
	}))
	hs.app.Use(ec_middleware.Recover())
	hs.app.Use(ec_middleware.Logger())
	hs.app.Use(middleware.MIME())
	if !hs.wc.devMode {

		hs.app.Use(middleware.CSP())
		hs.app.Use(middleware.HSTS())
	}

	return hs, nil
}

func (as *APIServer) registerEndpoints() error {

	// Blog

	blogCotroller := controllers.NewBlog(as.serv.blogStore, as.wc.HostName)
	as.app.GET("/blog", blogCotroller.ListBlogPosts)
	as.app.GET("/blog/:id", blogCotroller.GetBlogPost).Name = "BlogPost"

	// GQL
	if as.wc.enableGQL {
		gql, err := services.InitGQL(as.wc.devMode, as.serv.blogStore)
		if err != nil {
			return err
		}

		handler := controllers.GQLHandler(gql)
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
	return nil
}

func (as *APIServer) Serve() error {

	err := as.registerEndpoints()
	if err != nil {
		return err
	}
	wc := as.wc
	fmt.Println("webconf", wc)

	return as.app.Start(":" + strconv.Itoa(wc.HTTPPort))

}

func (as *APIServer) BaseAddr() string {

	return as.wc.HostName

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
		hs.wc.devMode = true
		return
	})
}

func WithGQL() Option {

	return newFuncOption(func(as *APIServer) (err error) {

		as.wc.enableGQL = true

		return
	})
}

func WithBlogStore(blogStore blog.BlogStore) Option {

	return newFuncOption(func(as *APIServer) (err error) {

		as.serv.blogStore = blogStore
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
