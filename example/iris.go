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
	api := iris.New()
	// define the api
	api.Get("/hello", func(ctx *iris.Context) {
		ctx.SetStatusCode(iris.StatusOK)
		ctx.SetBodyString("hello, world!")
	})

	api.PreListen(config.Server{ListeningAddr: ""})
	return api.ServeRequest
}
