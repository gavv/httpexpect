package httpexpect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

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

// the expErrorSuppressor is used as a Reporter to suppress an expected error
type expErrorSuppressor struct {
	backend          *assert.Assertions
	formatter        Formatter
	isExpError       isExpErrorFunc
	expErrorOccurred bool
}

type isExpErrorFunc func(err error) bool

func newExpErrorSuppressor(
	t assert.TestingT, isExpectedErr isExpErrorFunc,
) *expErrorSuppressor {
	return &expErrorSuppressor{
		backend:    assert.New(t),
		formatter:  &DefaultFormatter{},
		isExpError: isExpectedErr,
	}
}

func (h *expErrorSuppressor) Success(ctx *AssertionContext) {
}

func (h *expErrorSuppressor) Failure(
	ctx *AssertionContext, failure *AssertionFailure,
) {
	for _, e := range failure.Errors {
		if h.isExpError(e) {
			h.expErrorOccurred = true
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
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := WithConfig(Config{
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
	assert.True(t, suppressor.expErrorOccurred)
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
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := WithConfig(Config{
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
	assert.True(t, suppressor.expErrorOccurred)
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
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := WithConfig(Config{
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
	assert.True(t, suppressor.expErrorOccurred)
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
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := WithConfig(Config{
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
	assert.True(t, suppressor.expErrorOccurred)
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
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return strings.Contains(err.Error(), "context deadline exceeded")
		})
	e := WithConfig(Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	e.GET("/waitForPerRequestTimeout").
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, suppressor.expErrorOccurred)
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
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
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
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return strings.Contains(err.Error(), "context deadline exceeded")
		})
	e := WithConfig(Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	e.GET("/waitForPerRequestTimeout").
		WithContext(ctx).
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, suppressor.expErrorOccurred)
}

func TestE2EContext_PerRequestWithTimeoutCancelledByContext(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context deadline expected error
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return errors.Is(err, context.Canceled)
		})
	e := WithConfig(Config{
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
	assert.True(t, suppressor.expErrorOccurred)
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

		e := WithConfig(Config{
			BaseURL:          ts.URL,
			AssertionHandler: &mockAssertionHandler{},
		})

		e.GET("/").
			WithContext(ctx).
			WithMaxRetries(1).
			WithRetryPolicy(RetryAllErrors).
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
		suppressor := newExpErrorSuppressor(t,
			func(err error) bool {
				return errors.Is(err, context.Canceled)
			})

		e := WithConfig(Config{
			BaseURL:          ts.URL,
			AssertionHandler: suppressor,
		})

		e.GET("/").
			WithContext(ctx).
			WithMaxRetries(100).
			WithRetryPolicy(RetryAllErrors).
			WithRetryDelay(10*time.Millisecond, 10*time.Millisecond).
			Expect()

		<-isCtxCancelled

		assert.GreaterOrEqual(t, callCount, 2)
		assert.True(t, suppressor.expErrorOccurred)
	})
}
