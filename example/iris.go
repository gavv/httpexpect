package example

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/config"
	"github.com/valyala/fasthttp"
)

// IrisHandler creates fasthttp.RequestHandler for Iris web framework.
//
// Implemented API:
//  GET /hello            print "hello, world"
func IrisHandler() fasthttp.RequestHandler {
	iris.Get("/hello", func(ctx *iris.Context) {
		ctx.SetStatusCode(iris.StatusOK)
		ctx.SetBodyString("hello, world!")
	})
	
	return iris.NoListen().Handler
}
