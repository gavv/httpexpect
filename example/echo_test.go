package example

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestEcho_Standard(t *testing.T) {
	// create http.Handler
	handler := EchoHandlerStandard()

	// create httpexpect instance that will call htpp.Handler directly
	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client:   httpexpect.NewBinder(handler),
	})

	// run tests
	e.GET("/hello").
		Expect().
		Status(http.StatusOK).Body().Equal("hello, world!")
}

func TestEcho_Fast(t *testing.T) {
	// create fasthttp.RequestHandler
	handler := EchoHandlerFast()

	// create httpexpect instance that will call fasthtpp.RequestHandler directly
	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client:   httpexpect.NewFastBinder(handler),
	})

	// run tests
	e.GET("/hello").
		Expect().
		Status(http.StatusOK).Body().Equal("hello, world!")
}
