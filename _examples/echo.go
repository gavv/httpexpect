package example

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	echofasthttp "github.com/labstack/echo/engine/fasthttp"
	echostandard "github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"github.com/valyala/fasthttp"
)

// EchoServer creates HTTP server using echo framework.
//
// Implemented API:
//  GET /login             authenticate user and return JWT token
//  GET /restricted/hello  return "hello, world!" (requires authentication)
func EchoServer() *echo.Echo {
	e := echo.New()

	e.POST("/login", func(ctx echo.Context) error {
		username := ctx.FormValue("username")
		password := ctx.FormValue("password")

		if username == "ford" && password == "betelgeuse7" {
			// create token
			token := jwt.New(jwt.SigningMethodHS256)

			// generate encoded token and send it as response
			t, err := token.SignedString([]byte("seret"))
			if err != nil {
				return err
			}
			return ctx.JSON(http.StatusOK, map[string]string{
				"token": t,
			})
		}

		return echo.ErrUnauthorized
	})

	r := e.Group("/restricted")

	r.Use(middleware.JWT([]byte("seret")))

	r.GET("/hello", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "hello, world!")
	})

	return e
}

// EchoHandlerStandard creates http.Handler for EchoServer().
func EchoHandlerStandard() http.Handler {
	server := echostandard.New("")
	server.SetHandler(EchoServer())
	return http.Handler(server)
}

// EchoHandlerFast creates fasthttp.RequestHandler for EchoServer().
func EchoHandlerFast() fasthttp.RequestHandler {
	server := echofasthttp.New("")
	server.SetHandler(EchoServer())
	return func(ctx *fasthttp.RequestCtx) {
		server.ServeHTTP(ctx)
	}
}
