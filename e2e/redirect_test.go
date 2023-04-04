package e2e

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
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

func testRedirects(t *testing.T, createFn func(httpexpect.Reporter) *httpexpect.Expect) {
	t.Run("get301", func(t *testing.T) {
		e := createFn(httpexpect.NewAssertReporter(t))

		e.GET("/redirect301").
			WithRedirectPolicy(httpexpect.DontFollowRedirects).
			Expect().
			Status(http.StatusMovedPermanently)

		e.GET("/redirect301").
			WithRedirectPolicy(httpexpect.FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`default_response`)

		e.GET("/redirect301").
			WithRedirectPolicy(httpexpect.FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`default_response`)
	})

	t.Run("post301", func(t *testing.T) {
		e := createFn(httpexpect.NewAssertReporter(t))

		e.POST("/redirect301").
			WithText(`custom_response`).
			WithRedirectPolicy(httpexpect.DontFollowRedirects).
			Expect().
			Status(http.StatusMovedPermanently)

		e.POST("/redirect301").
			WithText(`custom_response`).
			WithRedirectPolicy(httpexpect.FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`default_response`)

		e.POST("/redirect301").
			WithText(`custom_response`).
			WithRedirectPolicy(httpexpect.FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`default_response`)
	})

	t.Run("get308", func(t *testing.T) {
		e := createFn(httpexpect.NewAssertReporter(t))

		e.GET("/redirect308").
			WithRedirectPolicy(httpexpect.DontFollowRedirects).
			Expect().
			Status(http.StatusPermanentRedirect)

		e.GET("/redirect308").
			WithRedirectPolicy(httpexpect.FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`default_response`)

		e.GET("/redirect308").
			WithRedirectPolicy(httpexpect.FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`default_response`)
	})

	t.Run("post308", func(t *testing.T) {
		e := createFn(httpexpect.NewAssertReporter(t))

		e.POST("/redirect308").
			WithText(`custom_response`).
			WithRedirectPolicy(httpexpect.DontFollowRedirects).
			Expect().
			Status(http.StatusPermanentRedirect)

		e.POST("/redirect308").
			WithText(`custom_response`).
			WithRedirectPolicy(httpexpect.FollowAllRedirects).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`custom_response`)

		e.POST("/redirect308").
			WithText(`custom_response`).
			WithRedirectPolicy(httpexpect.FollowRedirectsWithoutBody).
			Expect().
			Status(http.StatusPermanentRedirect)
	})

	t.Run("max redirects", func(t *testing.T) {
		t.Run("no max redirects set", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			e.POST("/double_redirect").
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})

		t.Run("max redirects above actual", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			e.POST("/double_redirect").
				WithMaxRedirects(2).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})

		t.Run("max redirects below actual", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			e.POST("/double_redirect").
				WithMaxRedirects(1).
				Expect().
				Status(http.StatusOK)

			assert.True(t, reporter.failed)
		})

		t.Run("max redirects equal to actual", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			e.POST("/redirect308").
				WithMaxRedirects(1).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})

		t.Run("max redirects zero with redirect", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			e.POST("/redirect308").
				WithMaxRedirects(0).
				Expect().
				Status(http.StatusOK)

			assert.True(t, reporter.failed)
		})

		t.Run("max redirects zero without redirect", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			e.POST("/content").
				WithMaxRedirects(0).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})
	})
}

func TestE2ERedirect_Live(t *testing.T) {
	handler := createRedirectHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testRedirects(t, func(rep httpexpect.Reporter) *httpexpect.Expect {
		return httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: rep,
		})
	})
}

func TestE2ERedirect_BinderStandard(t *testing.T) {
	handler := createRedirectHandler()

	testRedirects(t, func(rep httpexpect.Reporter) *httpexpect.Expect {
		return httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: httpexpect.NewBinder(handler),
			},
		})
	})
}

func TestE2ERedirect_BinderFast(t *testing.T) {
	handler := createRedirectFastHandler()

	testRedirects(t, func(rep httpexpect.Reporter) *httpexpect.Expect {
		return httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: httpexpect.NewFastBinder(handler),
			},
		})
	})
}
