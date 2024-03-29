package examples

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

// fastHTTPTester returns a new Expect instance to test FastHTTPHandler().
func fastHTTPTester(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		// Pass requests directly to FastHTTPHandler.
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(FastHTTPHandler()),
			Jar:       httpexpect.NewCookieJar(),
		},
		// Report errors using testify.
		Reporter: httpexpect.NewAssertReporter(t),
	})
}

func TestFastHTTP(t *testing.T) {
	e := fastHTTPTester(t)

	e.GET("/ping").Expect().
		Status(200).
		Text().IsEqual("pong")
}
