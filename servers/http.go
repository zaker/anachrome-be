package servers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/zaker/anachrome-be/controllers"
	"github.com/zaker/anachrome-be/middleware"
	"github.com/zaker/anachrome-be/services"

	"github.com/labstack/echo/v4"
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
	// e.Use(ec_middleware.CORSWithConfig(ec_middleware.CORSConfig{
	// 	// Skipper:      conf.DebugSkipper,
	// 	AllowOrigins: []string{conf.HostURI()},
	// 	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	// }))
	// hs.app.Use(ec_middleware.HTTPSRedirectWithConfig(ec_middleware.RedirectConfig{
	// 	Skipper: func(c echo.Context) bool {

	// 		if hs.isLocal || strings.HasPrefix(c.Request().URL.Path, "/.well-known/") {
	// 			return true
	// 		}
	// 		return false
	// 	},
	// }))

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

func WithOAuth2(oauthOpt OAuth2Option) Option {

	return newFuncOption(func(hs *APIServer) error {
		// sigKeySet, err := service.GetOIDCKeySet(oauthOpt.AuthServer)
		// if err != nil {
		// 	return fmt.Errorf("Couldn't get keyset: %v", err)
		// }

		// rsaJWTHandler := jwtmiddleware.New(jwtmiddleware.Config{
		// 	ValidationKeyGetter: func(t *jwt.Token) (interface{}, error) {

		// 		if t.Method.Alg() != "RS256" {
		// 			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		// 		}
		// 		return sigKeySet[t.Header["kid"].(string)], nil

		// 	},
		// 	ContextKey:    "user-jwt",
		// 	SigningMethod: jwt.SigningMethodRS256,
		// })

		// onRS256Pass := func(ctx irisCtx.Context, err error) {

		// 	if err == nil || err.Error() == "unexpected jwt signing method=RS256" {
		// 		return
		// 	}
		// 	jwtmiddleware.OnError(ctx, err)
		// }
		// hmacJWTHandler := jwtmiddleware.New(jwtmiddleware.Config{
		// 	ValidationKeyGetter: func(t *jwt.Token) (interface{}, error) {

		// 		if t.Method.Alg() != "HS256" {
		// 			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		// 		}
		// 		return oauthOpt.ApiSecret, nil
		// 	},
		// 	ContextKey:    "service-jwt",
		// 	SigningMethod: jwt.SigningMethodHS256,
		// 	ErrorHandler:  onRS256Pass,
		// })

		// if len(oauthOpt.Issuer) == 0 {
		// 	oauthOpt.Issuer = oauthOpt.AuthServer.String()
		// }

		// claimsHandler := claimsmiddleware.New(oauthOpt.Audience, oauthOpt.Issuer)

		// auth := func(ctx irisCtx.Context) {
		// 	hmacJWTHandler.Serve(ctx)
		// 	serviceToken := ctx.Values().Get("service-jwt")
		// 	if serviceToken == nil {
		// 		rsaJWTHandler.Serve(ctx)
		// 	}

		// }
		// hs.app.Use(auth)
		// hs.app.Use(claimsHandler.Validate)
		return nil
	})
}

func (hs *APIServer) registerEndpoints() {

	// hs.app.Static("/", conf.AppDir)
	// hs.app.GET("/", func(c echo.Context) (err error) {
	// 	pusher, ok := c.Response().Writer.(http.Pusher)
	// 	if ok {
	// 		for _, f := range s.PushFiles {
	// 			if err = pusher.Push(f, nil); err != nil {
	// 				return
	// 			}
	// 		}
	// 	}
	// 	return c.File(s.IndexPath)
	// })

	// Info
	hs.app.Any("/info", controllers.Info)
	// GQL

	gql, err := controllers.InitGQL(hs.devMode)
	if err != nil {
		log.Fatal(err)
	}
	hs.app.Any("/gql", echo.WrapHandler(gql.Handler()))

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
