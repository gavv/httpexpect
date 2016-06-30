package example

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestIris(t *testing.T) {
	// create fasthttp.RequestHandler
	handler := IrisHandler()

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
