package httpexpect

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// the waitHandler sleeps and fails (500) for a number of calls and succeeds
// (200) in the last one
type waitHandler struct {
	mux            *http.ServeMux
	callIdx        int
	callsDurations []time.Duration
	callCount      int
}

func (h *waitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *waitHandler) Wait(w http.ResponseWriter, r *http.Request) {
	h.callCount++
	i := h.callIdx
	if i <= len(h.callsDurations)-1 {
		d := h.callsDurations[i]
		time.Sleep(d)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func newWaitHandler(callsDurations []time.Duration) *waitHandler {
	mux := http.NewServeMux()

	handler := &waitHandler{
		mux:            mux,
		callsDurations: callsDurations,
	}

	mux.HandleFunc("/wait", handler.Wait)

	return handler
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

	handler := newWaitHandler([]time.Duration{
		time.Duration(3) * time.Second, // single call with wait 3 secs
	})

	cancelIn := time.Duration(2) * time.Second // cancel in 2 secs

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

	go func() {
		time.Sleep(cancelIn)
		cancel()
	}()

	e.GET("/wait").
		Expect()

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
}

func TestGlobalContextWithRetries(t *testing.T) {

	handler := newWaitHandler([]time.Duration{
		time.Duration(3) * time.Second, // first call will wait 3 secs
		time.Duration(3) * time.Second, // second call will wait 3 secs
	}) // third call will succeed

	cancelIn := time.Duration(5) * time.Second // cancel in 5 secs so that
	// the middle call is cancelled

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

	// cancel in cancelIn
	go func() {
		time.Sleep(cancelIn)
		cancel()
	}()

	// wait for waitFor
	e.GET("/wait").
		WithMaxRetries(3).
		Expect()

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
	// two calls; the third retry will NOT take place
	assert.Equal(t, 2, handler.callCount)
}

func TestPerRequestContext(t *testing.T) {

	handler := newWaitHandler([]time.Duration{
		time.Duration(3) * time.Second, // single call with wait 3 secs
	})

	cancelIn := time.Duration(2) * time.Second // cancel in 2 secs

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

	go func() {
		time.Sleep(cancelIn)
		cancel()
	}()

	e.GET("/wait").
		WithContext(ctx).
		Expect()

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
}

func TestPerRequestContextWithRetries(t *testing.T) {

	handler := newWaitHandler([]time.Duration{
		time.Duration(3) * time.Second, // first call will wait 3 secs
		time.Duration(3) * time.Second, // second call/first retry will
		// wait 3 secs
	}) // third call/second retry will
	// succeed

	cancelIn := time.Duration(5) * time.Second // cancel in 5 secs so that the
	// middle call is cancelled

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

	go func() {
		time.Sleep(cancelIn)
		cancel()
	}()

	e.GET("/wait").
		WithMaxRetries(2).
		WithContext(ctx).
		Expect()

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
	// two calls: the first one and the first retry; the second retry will NOT
	// take place
	assert.Equal(t, 2, handler.callCount)
}

func TestPerRequestWithTimeout(t *testing.T) {

	handler := newWaitHandler([]time.Duration{
		time.Duration(3) * time.Second, // single call will wait 3 secs
	})

	timeoutAt := time.Duration(2) * time.Second // time out at 2 secs

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

	e.GET("/wait").
		WithTimeout(timeoutAt).
		Expect()

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
	// one call; timed-out
	assert.Equal(t, 1, handler.callCount)
}

func TestPerRequestWithTimeoutAndWithRetries(t *testing.T) {

	handler := newWaitHandler([]time.Duration{
		time.Duration(3) * time.Second, // first call will wait 3 secs
		time.Duration(3) * time.Second, // second call/first retry will
		// wait 3 secs
	}) // third call/second retry will
	// not wait

	timeoutAt := time.Duration(2) * time.Second // timeout at 2 secs; both first
	// and second calls will time
	// out

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

	e.GET("/wait").
		WithTimeout(timeoutAt).
		WithMaxRetries(2).
		Expect().
		Status(200)

	// expected error should occur
	assert.True(t, reporter.expErrorOccurred)
	// three calls; the two timed-out and the one that normally returned
	assert.Equal(t, 3, handler.callCount)
}
