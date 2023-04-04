package e2e

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

const (
	TimeOutDuration = 500 * time.Millisecond
)

type waitHandler struct {
	mux           *http.ServeMux
	callCount     int
	retriesToFail int
	retriesDone   chan struct{}
	sync.RWMutex
}

func (h *waitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *waitHandler) waitForContextCancellation(w http.ResponseWriter, r *http.Request) {
	callCount := h.incrCallCount()
	// if retries-to-fail are not set then simply wait for the cancellation
	if h.retriesToFail == 0 {
		<-r.Context().Done()
	} else {
		// if retries-to-fail are set then make sure they are exhausted before
		// waiting for cancellation
		if callCount < h.retriesToFail+1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			h.retriesDone <- struct{}{}
			<-r.Context().Done()
		}
	}
}

func (h *waitHandler) waitForPerRequestTimeout(w http.ResponseWriter, r *http.Request) {
	callCount := h.incrCallCount()
	// if retries-to-fail are not set or not exhausted yet, simply wait for
	// the timeout
	if h.retriesToFail == 0 || callCount < h.retriesToFail+1 {
		<-r.Context().Done()
	} else {
		// otherwise succeed
		w.WriteHeader(http.StatusOK)
	}
}

func (h *waitHandler) incrCallCount() int {
	h.Lock()
	defer h.Unlock()
	h.callCount++
	r := h.callCount
	return r
}

func (h *waitHandler) getCallCount() int {
	h.RLock()
	defer h.RUnlock()
	return h.callCount
}

func newWaitHandler(retriesToFail int) *waitHandler {
	mux := http.NewServeMux()

	handler := &waitHandler{
		mux:           mux,
		retriesToFail: retriesToFail,
		retriesDone:   make(chan struct{}),
	}

	mux.HandleFunc("/waitForContextCancellation", handler.waitForContextCancellation)
	mux.HandleFunc("/waitForPerRequestTimeout", handler.waitForPerRequestTimeout)

	return handler
}

func (h *waitHandler) waitForRetries() {
	<-h.retriesDone
}

type errorSuppressor struct {
	backend               *assert.Assertions
	formatter             httpexpect.Formatter
	isExpectedError       func(err error) bool
	expectedErrorOccurred bool
}

func newErrorSuppressor(
	t assert.TestingT, isExpectedError func(err error) bool,
) *errorSuppressor {
	return &errorSuppressor{
		backend:         assert.New(t),
		formatter:       &httpexpect.DefaultFormatter{},
		isExpectedError: isExpectedError,
	}
}

func (h *errorSuppressor) Success(ctx *httpexpect.AssertionContext) {
}

func (h *errorSuppressor) Failure(
	ctx *httpexpect.AssertionContext, failure *httpexpect.AssertionFailure,
) {
	for _, e := range failure.Errors {
		if h.isExpectedError(e) {
			h.expectedErrorOccurred = true
			return
		}
	}
	h.backend.Fail(h.formatter.FormatFailure(ctx, failure))
}

func TestE2EContext_GlobalCancel(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	suppressor := newErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:          server.URL,
		Context:          ctx,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/waitForContextCancellation").
			Expect()
		done <- struct{}{}
	}()

	cancel()

	<-done

	// expected error should occur
	assert.True(t, suppressor.expectedErrorOccurred)
}

func TestE2EContext_GlobalWithRetries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	maxRetries := 3
	retriesToFail := 2
	handler := newWaitHandler(retriesToFail)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	suppressor := newErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:          server.URL,
		Context:          ctx,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/waitForContextCancellation").
			WithMaxRetries(maxRetries).
			Expect()
		done <- struct{}{}
	}()

	handler.waitForRetries() // wait for the retries-set-to-fail
	cancel()                 // cancel the rest

	<-done

	// expected error should occur
	assert.True(t, suppressor.expectedErrorOccurred)
	// first call + retries to fail should be the call count
	assert.Equal(t, retriesToFail+1, handler.getCallCount())
}

func TestE2EContext_PerRequest(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	suppressor := newErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/waitForContextCancellation").
			WithContext(ctx).
			Expect()
		done <- struct{}{}
	}()

	cancel()

	<-done

	// expected error should occur
	assert.True(t, suppressor.expectedErrorOccurred)
}

func TestE2EContext_PerRequestWithRetries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	maxRetries := 3
	retriesToFail := 2
	handler := newWaitHandler(retriesToFail)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	suppressor := newErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/waitForContextCancellation").
			WithMaxRetries(maxRetries).
			WithContext(ctx).
			Expect()
		done <- struct{}{}
	}()

	handler.waitForRetries() // wait for the retries-set-to-fail
	cancel()                 // cancel the rest

	<-done

	// expected error should occur
	assert.True(t, suppressor.expectedErrorOccurred)
	// first call + retries to fail should be the call count
	assert.Equal(t, retriesToFail+1, handler.getCallCount())
}

func TestE2EContext_PerRequestWithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	// config with context deadline expected error
	suppressor := newErrorSuppressor(t,
		func(err error) bool {
			return strings.Contains(err.Error(), "context deadline exceeded")
		})
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	e.GET("/waitForPerRequestTimeout").
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, suppressor.expectedErrorOccurred)
}

func TestE2EContext_PerRequestWithTimeoutAndRetries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	maxRetries := 3
	retriesToFail := 2
	handler := newWaitHandler(retriesToFail)

	server := httptest.NewServer(handler)
	defer server.Close()

	// this call will terminate with success
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
	})

	e.GET("/waitForPerRequestTimeout").
		WithTimeout(TimeOutDuration).
		WithMaxRetries(maxRetries).
		Expect().
		Status(http.StatusOK)

	// first call + retries to fail should be the call count
	assert.Equal(t, retriesToFail+1, handler.getCallCount())
}

func TestE2EContext_PerRequestWithTimeoutCancelledByTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context deadline expected error
	suppressor := newErrorSuppressor(t,
		func(err error) bool {
			return strings.Contains(err.Error(), "context deadline exceeded")
		})
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	e.GET("/waitForPerRequestTimeout").
		WithContext(ctx).
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, suppressor.expectedErrorOccurred)
}

func TestE2EContext_PerRequestWithTimeoutCancelledByContext(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context deadline expected error
	suppressor := newErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/waitForContextCancellation").
			WithContext(ctx).
			WithTimeout(TimeOutDuration).
			Expect()
		done <- struct{}{}
	}()

	cancel() // cancel the rest

	<-done

	// expected error should occur
	assert.True(t, suppressor.expectedErrorOccurred)
}

func TestE2EContext_PerRequestRetry(t *testing.T) {
	var m sync.Mutex

	t.Run("not cancelled", func(t *testing.T) {
		var callCount int

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.Lock()
			defer m.Unlock()
			callCount++

			if callCount > 1 {
				w.WriteHeader(http.StatusOK)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:  ts.URL,
			Reporter: t,
		})

		e.GET("/").
			WithContext(ctx).
			WithMaxRetries(1).
			WithRetryPolicy(httpexpect.RetryAllErrors).
			WithRetryDelay(time.Millisecond, time.Millisecond).
			Expect()

		assert.Equal(t, 2, callCount)
	})

	t.Run("cancelled after first retry attempt", func(t *testing.T) {
		var callCount int
		ctxCancellation := make(chan bool, 1) // To cancel context after first retry attempt
		isCtxCancelled := make(chan bool, 1)  // To wait until cancel() is called

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.Lock()
			defer m.Unlock()
			callCount++

			if callCount == 2 {
				ctxCancellation <- true
			}

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			<-ctxCancellation
			cancel()
			isCtxCancelled <- true
		}()

		// Config with context cancelled error
		suppressor := newErrorSuppressor(t,
			func(err error) bool {
				return errors.Is(err, context.Canceled)
			})

		e := httpexpect.WithConfig(httpexpect.Config{
			BaseURL:          ts.URL,
			AssertionHandler: suppressor,
		})

		e.GET("/").
			WithContext(ctx).
			WithMaxRetries(100).
			WithRetryPolicy(httpexpect.RetryAllErrors).
			WithRetryDelay(10*time.Millisecond, 10*time.Millisecond).
			Expect()

		<-isCtxCancelled

		assert.GreaterOrEqual(t, callCount, 2)
		assert.True(t, suppressor.expectedErrorOccurred)
	})
}
