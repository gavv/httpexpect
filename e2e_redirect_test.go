package httpexpect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valyala/fasthttp"
)

func createRedirectHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/foo", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`hello`))
	})

	mux.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/foo", http.StatusFound)
	})

	return mux
}

func createRedirectFastHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/foo":
			ctx.SetBody([]byte(`hello`))

		case "/bar":
			ctx.Redirect("/foo", http.StatusFound)
		}
	}
}

func testRedirectHandler(e *Expect) {
	e.POST("/bar").
		Expect().
		Status(http.StatusOK).Body().Equal(`hello`)
}

func TestE2ERedirectLive(t *testing.T) {
	handler := createRedirectHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testRedirectHandler(New(t, server.URL))
}

func TestE2ERedirectBinderStandard(t *testing.T) {
	handler := createRedirectHandler()

	testRedirectHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewBinder(handler),
		},
	}))
}

func TestE2ERedirectBinderFast(t *testing.T) {
	handler := createRedirectFastHandler()

	testRedirectHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewFastBinder(handler),
		},
	}))
}
