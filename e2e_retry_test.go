package httpexpect

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/valyala/fasthttp"
)

type retryController struct {
	mu         sync.Mutex
	countback  int
	statuscode int
}

func (rc *retryController) Reset(countback, statuscode int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.countback = countback
	rc.statuscode = statuscode
}

func (rc *retryController) GetStatusCode() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.countback > 0 {
		rc.countback--
		return rc.statuscode
	}
	return http.StatusOK
}

type transportController struct {
	http.RoundTripper

	mu        sync.Mutex
	countback int
}

func (tc *transportController) Reset(countback int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.countback = countback
}

func (tc *transportController) RoundTrip(req *http.Request) (*http.Response, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.countback != 0 {
		tc.countback--
		return nil, temporaryNetworkErr{}
	}

	return tc.RoundTripper.RoundTrip(req)
}

type temporaryNetworkErr struct {
	error
}

func (temporaryNetworkErr) Timeout() bool {
	return false
}

func (temporaryNetworkErr) Temporary() bool {
	return true
}

func createRetryHandler(rc *retryController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)

		w.WriteHeader(rc.GetStatusCode())
		_, _ = w.Write(b)
	})

	return mux
}

func createRetryFastHandler(rc *retryController) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/test":
			ctx.SetStatusCode(rc.GetStatusCode())
			ctx.SetBody(ctx.Request.Body())
		}
	}
}

func testRetries(
	t *testing.T,
	rc *retryController,
	tc *transportController,
	createFn func(Reporter) *Expect,
) {
	t.Run("MaxRetries", func(t *testing.T) {
		e := createFn(NewAssertReporter(t))

		rc.Reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			Expect().
			Status(http.StatusInternalServerError).Body().Equal(`test`)

		rc.Reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			WithMaxRetries(0).
			Expect().
			Status(http.StatusInternalServerError).Body().Equal(`test`)

		rc.Reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			WithMaxRetries(1).
			Expect().
			Status(http.StatusInternalServerError).Body().Equal(`test`)

		rc.Reset(2, http.StatusInternalServerError)
		e.POST("/test").
			WithText(`test`).
			WithMaxRetries(2).
			Expect().
			Status(http.StatusOK).Body().Equal(`test`)
	})

	t.Run("DontRetry", func(t *testing.T) {
		e := createFn(newMockReporter(t))

		rc.Reset(1, http.StatusInternalServerError)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(DontRetry).
			Expect().
			Status(http.StatusInternalServerError).chain.assertOK(t)

		rc.Reset(1, http.StatusBadRequest)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(DontRetry).
			Expect().
			Status(http.StatusBadRequest).chain.assertOK(t)

		rc.Reset(0, http.StatusOK)
		tc.Reset(1)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(DontRetry).
			Expect().chain.assertFailed(t)
	})

	t.Run("RetryTemporaryNetworkErrors", func(t *testing.T) {
		e := createFn(newMockReporter(t))

		rc.Reset(1, http.StatusInternalServerError)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryTemporaryNetworkErrors).
			Expect().
			Status(http.StatusInternalServerError).chain.assertOK(t)

		rc.Reset(1, http.StatusBadRequest)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryTemporaryNetworkErrors).
			Expect().
			Status(http.StatusBadRequest).chain.assertOK(t)

		rc.Reset(0, http.StatusOK)
		tc.Reset(1)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryTemporaryNetworkErrors).
			Expect().
			Status(http.StatusOK).chain.assertOK(t)
	})

	t.Run("RetryTemporaryNetworkAndServerErrors", func(t *testing.T) {
		e := createFn(newMockReporter(t))

		rc.Reset(1, http.StatusInternalServerError)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryTemporaryNetworkAndServerErrors).
			Expect().
			Status(http.StatusOK).chain.assertOK(t)

		rc.Reset(1, http.StatusBadRequest)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryTemporaryNetworkAndServerErrors).
			Expect().
			Status(http.StatusBadRequest).chain.assertOK(t)

		rc.Reset(0, http.StatusOK)
		tc.Reset(1)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryTemporaryNetworkAndServerErrors).
			Expect().
			Status(http.StatusOK).chain.assertOK(t)
	})

	t.Run("RetryAllErrors", func(t *testing.T) {
		e := createFn(newMockReporter(t))

		rc.Reset(1, http.StatusInternalServerError)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryAllErrors).
			Expect().
			Status(http.StatusOK).chain.assertOK(t)

		rc.Reset(1, http.StatusBadRequest)
		tc.Reset(0)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryAllErrors).
			Expect().
			Status(http.StatusOK).chain.assertOK(t)

		rc.Reset(0, http.StatusOK)
		tc.Reset(1)
		e.POST("/test").
			WithMaxRetries(1).WithRetryPolicy(RetryAllErrors).
			Expect().
			Status(http.StatusOK).chain.assertOK(t)
	})
}

func TestE2ERetryLive(t *testing.T) {
	rc := &retryController{}

	handler := createRetryHandler(rc)

	tc := &transportController{RoundTripper: http.DefaultTransport}

	server := httptest.NewServer(handler)
	defer server.Close()

	testRetries(t, rc, tc, func(rep Reporter) *Expect {
		return WithConfig(Config{
			BaseURL:  server.URL,
			Reporter: rep,
			Client: &http.Client{
				Transport: tc,
			},
		})
	})
}

func TestE2ERetryBinderStandard(t *testing.T) {
	rc := &retryController{}

	handler := createRetryHandler(rc)

	tc := &transportController{RoundTripper: NewBinder(handler)}

	testRetries(t, rc, tc, func(rep Reporter) *Expect {
		return WithConfig(Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: tc,
			},
		})
	})
}

func TestE2ERetryBinderFast(t *testing.T) {
	rc := &retryController{}

	handler := createRetryFastHandler(rc)

	tc := &transportController{RoundTripper: NewFastBinder(handler)}

	testRetries(t, rc, tc, func(rep Reporter) *Expect {
		return WithConfig(Config{
			BaseURL:  "http://example.com",
			Reporter: rep,
			Client: &http.Client{
				Transport: tc,
			},
		})
	})
}
