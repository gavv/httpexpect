package example

import (
	"github.com/labstack/echo"
	echofast "github.com/labstack/echo/engine/fasthttp"
	echostandard "github.com/labstack/echo/engine/standard"
	"github.com/valyala/fasthttp"
	"net/http"
)

// EchoServer creates HTTP server using echo framework.
//
// Implemented API:
//  GET /hello            print "hello, world"
func EchoServer() *echo.Echo {
	ec := echo.New()

	ec.GET("/hello", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "hello, world!")
	})

	return ec
}

// EchoHandlerStandard creates http.Handler for EchoServer().
func EchoHandlerStandard() http.Handler {
	server := echostandard.New("")
	server.SetHandler(EchoServer())
	return http.Handler(server)
}

// EchoHandlerFast creates fasthttp.RequestHandler for EchoServer().
func EchoHandlerFast() fasthttp.RequestHandler {
	server := echofast.New("")
	server.SetHandler(EchoServer())
	return func(ctx *fasthttp.RequestCtx) {
		server.ServeHTTP(ctx)
	}
}
