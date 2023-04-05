package e2e

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type retryController struct {
	mu         sync.Mutex
	countback  int
	statuscode int
}

func (rc *retryController) getStatus() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.countback > 0 {
		rc.countback--
		return rc.statuscode
	}
	return http.StatusOK
}

func (rc *retryController) reset(countback, statuscode int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.countback = countback
	rc.statuscode = statuscode
}

type transportController struct {
	http.RoundTripper
	mu        sync.Mutex
	countback int
}

func (tc *transportController) RoundTrip(req *http.Request) (*http.Response, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.countback != 0 {
		tc.countback--
		return nil, timeoutNetworkErr{}
	}

	return tc.RoundTripper.RoundTrip(req)
}

func (tc *transportController) reset(countback int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.countback = countback
}

type timeoutNetworkErr struct {
	error
}

func (timeoutNetworkErr) Timeout() bool {
	return true
}

func (timeoutNetworkErr) Temporary() bool {
	return true
}

func createRetryHandler(rc *retryController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)

		w.WriteHeader(rc.getStatus())
		_, _ = w.Write(b)
	})

	return mux
}

func createRetryFastHandler(rc *retryController) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/test":
			ctx.SetStatusCode(rc.getStatus())
			ctx.SetBody(ctx.Request.Body())
		}
	}
}

func testRetries(
	t *testing.T,
	rc *retryController,
	tc *transportController,
	createFn func(httpexpect.Reporter) *httpexpect.Expect,
) {
	t.Run("MaxRetries", func(t *testing.T) {
		e := createFn(httpexpect.NewAssertReporter(t))

		rc.reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			Expect().
			Status(http.StatusInternalServerError).Body().IsEqual(`test`)

		rc.reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			WithMaxRetries(0).
			Expect().
			Status(http.StatusInternalServerError).Body().IsEqual(`test`)

		rc.reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			WithMaxRetries(1).
			Expect().
			Status(http.StatusInternalServerError).Body().IsEqual(`test`)

		rc.reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			WithMaxRetries(2).
			Expect().
			Status(http.StatusOK).Body().IsEqual(`test`)
	})

	t.Run("DontRetry", func(t *testing.T) {
		t.Run("dont retry internal server error", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusInternalServerError)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.DontRetry).
				Expect().
				Status(http.StatusInternalServerError)

			assert.False(t, reporter.failed)
		})

		t.Run("dont retry bad request", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusBadRequest)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.DontRetry).
				Expect().
				Status(http.StatusBadRequest)

			assert.False(t, reporter.failed)
		})

		t.Run("dont retry timeout error", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(0, http.StatusOK)
			tc.reset(1)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.DontRetry).
				Expect()

			assert.True(t, reporter.failed)
		})
	})

	t.Run("RetryTimeoutErrors", func(t *testing.T) {
		t.Run("dont retry internal server error", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusInternalServerError)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryTimeoutErrors).
				Expect().
				Status(http.StatusInternalServerError)

			assert.False(t, reporter.failed)
		})

		t.Run("dont retry bad request", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusBadRequest)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryTimeoutErrors).
				Expect().
				Status(http.StatusBadRequest)

			assert.False(t, reporter.failed)
		})

		t.Run("retry timeout error", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(0, http.StatusOK)
			tc.reset(1)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryTimeoutErrors).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})
	})

	t.Run("RetryTimeoutAndServerErrors", func(t *testing.T) {
		t.Run("retry internal server", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusInternalServerError)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryTimeoutAndServerErrors).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})

		t.Run("retry bad request", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusBadRequest)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryTimeoutAndServerErrors).
				Expect().
				Status(http.StatusBadRequest)

			assert.False(t, reporter.failed)
		})

		t.Run("retry timeout error", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(0, http.StatusOK)
			tc.reset(1)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryTimeoutAndServerErrors).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})
	})

	t.Run("RetryAllErrors", func(t *testing.T) {
		t.Run("retry internal server error", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusInternalServerError)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryAllErrors).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})

		t.Run("retry bad request", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(1, http.StatusBadRequest)
			tc.reset(0)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryAllErrors).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})

		t.Run("retry timeout error", func(t *testing.T) {
			reporter := &mockReporter{}
			e := createFn(reporter)

			rc.reset(0, http.StatusOK)
			tc.reset(1)
			e.POST("/test").
				WithMaxRetries(1).WithRetryPolicy(httpexpect.RetryAllErrors).
				Expect().
				Status(http.StatusOK)

			assert.False(t, reporter.failed)
		})
	})
}

func TestE2ERetry_Live(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	rc := &retryController{}

	handler := createRetryHandler(rc)

	tc := &transportController{RoundTripper: http.DefaultTransport}

	server := httptest.NewServer(handler)
	defer server.Close()

	testRetries(t, rc, tc, func(rep httpexpect.Reporter) *httpexpect.Expect {
		return httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  server.URL,
			Reporter: rep,
			Client: &http.Client{
				Transport: tc,
			},
		})
	})
}

func TestE2ERetry_BinderStandard(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	rc := &retryController{}

	handler := createRetryHandler(rc)

	tc := &transportController{RoundTripper: httpexpect.NewBinder(handler)}

	testRetries(t, rc, tc, func(rep httpexpect.Reporter) *httpexpect.Expect {
		return httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: tc,
			},
		})
	})
}

func TestE2ERetry_BinderFast(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	rc := &retryController{}

	handler := createRetryFastHandler(rc)

	tc := &transportController{RoundTripper: httpexpect.NewFastBinder(handler)}

	testRetries(t, rc, tc, func(rep httpexpect.Reporter) *httpexpect.Expect {
		return httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: tc,
			},
		})
	})
}
