package httpexpect

import (
	"context"
	"fmt"
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
	isExpError       isExpErrorFunc
	expErrorOccurred bool
}

type isExpErrorFunc func(message string, args ...interface{}) bool

func newExpErrorSuppressor(t assert.TestingT,
	isExpectedErr isExpErrorFunc) *expErrorSuppressor {
	return &expErrorSuppressor{backend: assert.New(t),
		isExpError: isExpectedErr}
}

func (r *expErrorSuppressor) Errorf(message string, args ...interface{}) {
	if !r.isExpError(message, args) {
		r.backend.Fail(fmt.Sprintf(message, args...))
	} else {
		r.expErrorOccurred = true
	}
}

func TestGlobalContextCancel(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	reporter := newExpErrorSuppressor(t,
		func(message string, args ...interface{}) bool {
			return strings.HasSuffix(message, "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: reporter,
		Context:  ctx,
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
	assert.True(t, reporter.expErrorOccurred)
}

func TestGlobalContextWithRetries(t *testing.T) {
	maxRetries := 3
	retriesToFail := 2
	handler := newWaitHandler(retriesToFail)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	reporter := newExpErrorSuppressor(t,
		func(message string, args ...interface{}) bool {
			return strings.HasSuffix(message, "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: reporter,
		Context:  ctx,
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
	assert.True(t, reporter.expErrorOccurred)
	// first call + retries to fail should be the call count
	assert.Equal(t, retriesToFail+1, handler.GetCallCount())
}

func TestPerRequestContext(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	reporter := newExpErrorSuppressor(t,
		func(message string, args ...interface{}) bool {
			return strings.HasSuffix(message, "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: reporter,
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
	assert.True(t, reporter.expErrorOccurred)
}

func TestPerRequestContextWithRetries(t *testing.T) {
	maxRetries := 3
	retriesToFail := 2
	handler := newWaitHandler(retriesToFail)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context cancel suppression
	reporter := newExpErrorSuppressor(t,
		func(message string, args ...interface{}) bool {
			return strings.HasSuffix(message, "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: reporter,
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
	assert.True(t, reporter.expErrorOccurred)
	// first call + retries to fail should be the call count
	assert.Equal(t, retriesToFail+1, handler.GetCallCount())
}

func TestPerRequestWithTimeout(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	// config with context deadline expected error
	reporter := newExpErrorSuppressor(t,
		func(message string, args ...interface{}) bool {
			return strings.HasSuffix(message, "context deadline exceeded")
		})
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: reporter,
	})

	e.GET("/WaitForPerRequestTimeout").
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
}

func TestPerRequestWithTimeoutAndWithRetries(t *testing.T) {
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

func TestPerRequestWithContextAndTimeout_CancelledByTimeout(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context deadline expected error
	reporter := newExpErrorSuppressor(t,
		func(message string, args ...interface{}) bool {
			return strings.HasSuffix(message, "context deadline exceeded")
		})
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: reporter,
	})

	e.GET("/WaitForPerRequestTimeout").
		WithContext(ctx).
		WithTimeout(TimeOutDuration).
		Expect()

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
}

func TestPerRequestWithContextAndTimeout_CancelledByContext(t *testing.T) {
	handler := newWaitHandler(0)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config with context deadline expected error
	reporter := newExpErrorSuppressor(t,
		func(message string, args ...interface{}) bool {
			return strings.HasSuffix(message, "context canceled")
		})
	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: reporter,
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
	assert.True(t, reporter.expErrorOccurred)
}
