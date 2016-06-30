package example

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestEchoStandard(t *testing.T) {
	// create http.Handler
	handler := EchoHandlerStandard()

	// create httpexpect instance that will call htpp.Handler directly
	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
	})

	// run tests
	e.GET("/hello").
		Expect().
		Status(http.StatusOK).Body().Equal("hello, world!")
}

func TestEchoFast(t *testing.T) {
	// create fasthttp.RequestHandler
	handler := EchoHandlerFast()

	// create httpexpect instance that will call fasthtpp.RequestHandler directly
	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
	})

	// run tests
	e.GET("/hello").
		Expect().
		Status(http.StatusOK).Body().Equal("hello, world!")
}
