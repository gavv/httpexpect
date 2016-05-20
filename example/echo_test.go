package example

import (
	"github.com/gavv/httpexpect"
	"github.com/gavv/httpexpect/fasthttpexpect"
	"net/http"
	"testing"
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
		Client:   fasthttpexpect.NewBinder(handler),
	})

	// run tests
	e.GET("/hello").
		Expect().
		Status(http.StatusOK).Body().Equal("hello, world!")
}
