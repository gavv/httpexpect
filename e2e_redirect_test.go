package httpexpect

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valyala/fasthttp"
)

func createRedirectHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/content", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, _ = w.Write([]byte(`default_response`))
		} else {
			b, _ := ioutil.ReadAll(r.Body)
			_, _ = w.Write(b)
		}
	})

	mux.HandleFunc("/redirect301", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/content", http.StatusMovedPermanently)
	})

	mux.HandleFunc("/redirect308", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/content", http.StatusPermanentRedirect)
	})

	mux.HandleFunc("/double_redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect308", http.StatusTemporaryRedirect)
	})

	return mux
}

func createRedirectFastHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/content":
			if ctx.IsGet() {
				ctx.SetBody([]byte(`default_response`))
			} else {
				ctx.SetBody(ctx.Request.Body())
			}

		case "/redirect301":
			ctx.Redirect("/content", http.StatusMovedPermanently)

		case "/redirect308":
			ctx.Redirect("/content", http.StatusPermanentRedirect)

		case "/double_redirect":
			ctx.Redirect("/redirect308", http.StatusTemporaryRedirect)
		}
	}
}

func testRedirects(t *testing.T, createFn func(Reporter) *Expect) {
	t.Run("get301", func(t *testing.T) {
		e := createFn(NewAssertReporter(t))

		e.GET("/redirect301").
			WithRedirectPolicy(DontFollowRedirects).
			Expect().
			Status(http.StatusMovedPermanently)

		e.GET("/redirect301").
			WithRedirectPolicy(FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().Equal(`default_response`)

		e.GET("/redirect301").
			WithRedirectPolicy(FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusOK).Body().Equal(`default_response`)
	})

	t.Run("post301", func(t *testing.T) {
		e := createFn(NewAssertReporter(t))

		e.POST("/redirect301").
			WithText(`custom_response`).
			WithRedirectPolicy(DontFollowRedirects).
			Expect().
			Status(http.StatusMovedPermanently)

		e.POST("/redirect301").
			WithText(`custom_response`).
			WithRedirectPolicy(FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().Equal(`default_response`)

		e.POST("/redirect301").
			WithText(`custom_response`).
			WithRedirectPolicy(FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusOK).Body().Equal(`default_response`)
	})

	t.Run("get308", func(t *testing.T) {
		e := createFn(NewAssertReporter(t))

		e.GET("/redirect308").
			WithRedirectPolicy(DontFollowRedirects).
			Expect().
			Status(http.StatusPermanentRedirect)

		e.GET("/redirect308").
			WithRedirectPolicy(FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().Equal(`default_response`)

		e.GET("/redirect308").
			WithRedirectPolicy(FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusOK).Body().Equal(`default_response`)
	})

	t.Run("post308", func(t *testing.T) {
		e := createFn(NewAssertReporter(t))

		e.POST("/redirect308").
			WithText(`custom_response`).
			WithRedirectPolicy(DontFollowRedirects).
			Expect().
			Status(http.StatusPermanentRedirect)

		e.POST("/redirect308").
			WithText(`custom_response`).
			WithRedirectPolicy(FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().Equal(`custom_response`)

		e.POST("/redirect308").
			WithText(`custom_response`).
			WithRedirectPolicy(FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusPermanentRedirect)
	})

	t.Run("max-redirects", func(t *testing.T) {
		e := createFn(newMockReporter(t))

		e.POST("/double_redirect").
			Expect().chain.assertOK(t)

		e.POST("/double_redirect").
			WithMaxRedirects(2).
			Expect().chain.assertOK(t)

		e.POST("/double_redirect").
			WithMaxRedirects(1).
			Expect().chain.assertFailed(t)

		e.POST("/redirect308").
			WithMaxRedirects(1).
			Expect().chain.assertOK(t)

		e.POST("/redirect308").
			WithMaxRedirects(0).
			Expect().chain.assertFailed(t)

		e.POST("/content").
			WithMaxRedirects(0).
			Expect().chain.assertOK(t)
	})
}

func TestE2ERedirectLive(t *testing.T) {
	handler := createRedirectHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testRedirects(t, func(rep Reporter) *Expect {
		return WithConfig(Config{
			BaseURL:  server.URL,
			Reporter: rep,
		})
	})
}

func TestE2ERedirectBinderStandard(t *testing.T) {
	handler := createRedirectHandler()

	testRedirects(t, func(rep Reporter) *Expect {
		return WithConfig(Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: NewBinder(handler),
			},
		})
	})
}

func TestE2ERedirectBinderFast(t *testing.T) {
	handler := createRedirectFastHandler()

	testRedirects(t, func(rep Reporter) *Expect {
		return WithConfig(Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: NewFastBinder(handler),
			},
		})
	})
}
