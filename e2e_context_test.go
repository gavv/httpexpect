package httpexpect

import (
	"context"
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

func (h *waitHandler) WaitForContextCancellation(w http.ResponseWriter, r *http.Request) {
	callCount := h.IncrCallCount()
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

func (h *waitHandler) WaitForPerRequestTimeout(w http.ResponseWriter, r *http.Request) {
	callCount := h.IncrCallCount()
	// if retries-to-fail are not set or not exhausted yet, simply wait for
	// the timeout
	if h.retriesToFail == 0 || callCount < h.retriesToFail+1 {
		<-r.Context().Done()
	} else {
		// otherwise succeed
		w.WriteHeader(http.StatusOK)
	}
}

func (h *waitHandler) IncrCallCount() int {
	h.Lock()
	defer h.Unlock()
	h.callCount++
	r := h.callCount
	return r
}

func (h *waitHandler) GetCallCount() int {
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

	mux.HandleFunc("/WaitForContextCancellation", handler.WaitForContextCancellation)
	mux.HandleFunc("/WaitForPerRequestTimeout", handler.WaitForPerRequestTimeout)

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

func TestContextGlobalCancel(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return strings.Contains(err.Error(), "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:          server.URL,
		Context:          ctx,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/WaitForContextCancellation").
			Expect()
		done <- struct{}{}
	}()

	cancel()

	<-done

	// expected error should occur
	assert.True(t, suppressor.expErrorOccurred)
}

func TestContextGlobalWithRetries(t *testing.T) {
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
			return strings.Contains(err.Error(), "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:          server.URL,
		Context:          ctx,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/WaitForContextCancellation").
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
	assert.Equal(t, retriesToFail+1, handler.GetCallCount())
}

func TestContextPerRequest(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return strings.Contains(err.Error(), "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/WaitForContextCancellation").
			WithContext(ctx).
			Expect()
		done <- struct{}{}
	}()

	cancel()

	<-done

	// expected error should occur
	assert.True(t, suppressor.expErrorOccurred)
}

func TestContextPerRequestWithRetries(t *testing.T) {
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
			return strings.Contains(err.Error(), "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/WaitForContextCancellation").
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
	assert.Equal(t, retriesToFail+1, handler.GetCallCount())
}

func TestContextPerRequestWithTimeout(t *testing.T) {
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

	e.GET("/WaitForPerRequestTimeout").
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, suppressor.expErrorOccurred)
}

func TestContextPerRequestWithTimeoutAndRetries(t *testing.T) {
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

	e.GET("/WaitForPerRequestTimeout").
		WithTimeout(TimeOutDuration).
		WithMaxRetries(maxRetries).
		Expect().
		Status(http.StatusOK)

	// first call + retries to fail should be the call count
	assert.Equal(t, retriesToFail+1, handler.GetCallCount())
}

func TestContextPerRequestWithTimeoutCancelledByTimeout(t *testing.T) {
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

	e.GET("/WaitForPerRequestTimeout").
		WithContext(ctx).
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, suppressor.expErrorOccurred)
}

func TestContextPerRequestWithTimeoutCancelledByContext(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context deadline expected error
	suppressor := newExpErrorSuppressor(t,
		func(err error) bool {
			return strings.Contains(err.Error(), "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:          server.URL,
		AssertionHandler: suppressor,
	})

	done := make(chan struct{})

	go func() {
		e.GET("/WaitForContextCancellation").
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
