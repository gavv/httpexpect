// Package httpexpect helps to write nice tests for your HTTP API.
//
// Usage example
//
// See example directory:
//  - https://godoc.org/github.com/gavv/httpexpect/example
//  - https://github.com/gavv/httpexpect/tree/master/example
//
// Value equality
//
// Whenever values are checked for equality in httpexpect, they are converted
// to "canonical form":
//  - type aliases are removed
//  - numeric types are converted to float64
//  - non-nil interfaces pointing to nil slices and maps are replaced with nil interfaces
//  - structs are converted to map[string]interface{}
//
// This is equivalent to subsequently json.Marshal() and json.Unmarshal() the value.
//
// Failure handling
//
// When some check fails, failure is reported. If non-fatal failures are used
// (see Checker interface), execution is continued and instance that was checked
// is marked as failed.
//
// If specific instance is marked as failed, all subsequent checks are ignored
// for this instance and for any child instances retrieved after failure.
//
// Example:
//  array := NewArray(NewAssertChecker(t), []interface{}{"foo", 123})
//
//  e0 := array.Element(0)  // success
//  e1 := array.Element(1)  // success
//
//  s0 := e0.String()       // success
//  s1 := e1.String()       // failure; e1 and s1 are marked as failed, e0 and s0 are not
//
//  s0.Equal("foo")         // success
//  s1.Equal("bar")         // this check is ignored because s1 is marked as failed
package httpexpect

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Expect is a toplevel object that contains user Config and allows
// to construct Request objects.
type Expect struct {
	config Config
}

// Config contains various settings.
type Config struct {
	// BaseURL is a URL to prepended to all request. My be empty. If
	// non-empty, trailing slash is allowed but not required and is
	// appended automatically.
	BaseURL string

	// Client is used to send http.Request and receive http.Response.
	// Should not be nil.
	//
	// You can use http.DefaultClient or provide custom implementation.
	Client Client

	// Checker is used to compare arbitrary values and report failures.
	// Should not be nil.
	//
	// You can use AssertChecker or RequireChecker, or provide custom
	// implementation.
	Checker Checker

	// Logger is used to report various events.
	// May be nil.
	//
	// You can use CompactLogger or DebugLogger, or provide custom
	// implementation.
	//
	// You can also use CompactLogger or DebugLogger with alternative
	// LoggerBackend if you're happy with log format, but don't want
	// to send logs to testing.T.
	Logger Logger
}

// Client is used to send http.Request and receive http.Response.
type Client interface {
	// Do sends request and returns response.
	Do(*http.Request) (*http.Response, error)
}

// Checker is used to compare arbitrary values and report failures.
//
// We have to builtin implementations:
//  - AssertChecker (uses testify/assert)
//  - RequireChecker (uses testify/require)
type Checker interface {
	// Clone returns a copy of this checker instance. When Fail() is called,
	// it should affect only given instance and its *future* copies, but not
	// other copies.
	Clone() Checker

	// Failed checks if some previous check was failed. Clones inherit their
	// failed state from original instance. However, if a clone becomes failed
	// later, it doesn't affect original instance.
	Failed() bool

	// Fail reports failure.
	Fail(message string, args ...interface{})
}

// Logger is used to report various events.
type Logger interface {
	// Request is called before request is sent.
	Request(*http.Request)

	// Response is called after response is received.
	Response(*http.Response)
}

// LoggerBackend is log output interface.
type LoggerBackend interface {
	// Logf writes message to log.
	// Note that testing.T implements this interface.
	Logf(fmt string, args ...interface{})
}

// New returns a new Expect object.
//
// baseURL specifies URL to prepended to all request. My be empty. If non-empty,
// trailing slash is allowed but not required and is appended automatically.
//
// New is shorthand for WithConfig. It uses:
//  - http.DefaultClient as Client
//  - AssertChecker as Checker
//  - CompactLogger as Logger, with testing.T as LoggerBackend
//
// Example:
//  func TestAPI(t *testing.T) {
//      e := httpexpect.New(t, "http://example.org/")
//      e.GET("/path").Expect().Status(http.StatusOK)
//  }
func New(t *testing.T, baseURL string) *Expect {
	return WithConfig(Config{
		BaseURL: baseURL,
		Checker: NewAssertChecker(t),
		Logger:  NewCompactLogger(t),
	})
}

// WithConfig returns a new Expect object with given config.
//
// If Config.Client is nil, http.DefaultClient is used.
//
// Example:
//  func TestAPI(t *testing.T) {
//      e := httpexpect.WithConfig(httpexpect.Config{
//          BaseURL: "http://example.org/",
//          Client:  http.DefaultClient,
//          Checker: httpexpect.NewAssertChecker(t),
//          Logger:  httpexpect.NewDebugLogger(t),
//      })
//      e.GET("/path").Expect().Status(http.StatusOK)
//  }
func WithConfig(config Config) *Expect {
	if config.Client == nil {
		config.Client = http.DefaultClient
	}
	if config.Checker == nil {
		panic("config.Checker is nil")
	}
	return &Expect{config}
}

// Request returns a new Request object given HTTP method and url. Returned
// object allows to build request incrementally and send it to server.
//
// method specifies the HTTP method (GET, POST, PUT, etc.).
// url and args are passed to fmt.Sprintf().
// Config.BaseURL is prepended to final url.
func (e *Expect) Request(method, url string, args ...interface{}) *Request {
	config := e.config
	config.Checker = config.Checker.Clone()
	url = concatURLs(config.BaseURL, fmt.Sprintf(url, args...))
	return NewRequest(config, method, url)
}

func concatURLs(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if strings.HasSuffix(a, "/") {
		a = a[:len(a)-1]
	}
	if strings.HasPrefix(b, "/") {
		b = b[1:]
	}
	return a + "/" + b
}

// OPTIONS is a shorthand for Request("OPTIONS", url, args...).
func (e *Expect) OPTIONS(url string, args ...interface{}) *Request {
	return e.Request("OPTIONS", url, args...)
}

// HEAD is a shorthand for Request("HEAD", url, args...).
func (e *Expect) HEAD(url string, args ...interface{}) *Request {
	return e.Request("HEAD", url, args...)
}

// GET is a shorthand for Request("GET", url, args...).
func (e *Expect) GET(url string, args ...interface{}) *Request {
	return e.Request("GET", url, args...)
}

// POST is a shorthand for Request("POST", url, args...).
func (e *Expect) POST(url string, args ...interface{}) *Request {
	return e.Request("POST", url, args...)
}

// PUT is a shorthand for Request("PUT", url, args...).
func (e *Expect) PUT(url string, args ...interface{}) *Request {
	return e.Request("PUT", url, args...)
}

// PATCH is a shorthand for Request("PATCH", url, args...).
func (e *Expect) PATCH(url string, args ...interface{}) *Request {
	return e.Request("PATCH", url, args...)
}

// DELETE is a shorthand for Request("DELETE", url, args...).
func (e *Expect) DELETE(url string, args ...interface{}) *Request {
	return e.Request("DELETE", url, args...)
}

// Value is a shorthand for NewValue(Config.Checker.Clone(), value).
func (e *Expect) Value(value interface{}) *Value {
	return NewValue(e.config.Checker.Clone(), value)
}

// Object is a shorthand for NewObject(Config.Checker.Clone(), value).
func (e *Expect) Object(value map[string]interface{}) *Object {
	return NewObject(e.config.Checker.Clone(), value)
}

// Array is a shorthand for NewArray(Config.Checker.Clone(), value).
func (e *Expect) Array(value []interface{}) *Array {
	return NewArray(e.config.Checker.Clone(), value)
}

// String is a shorthand for NewString(Config.Checker.Clone(), value).
func (e *Expect) String(value string) *String {
	return NewString(e.config.Checker.Clone(), value)
}

// Number is a shorthand for NewNumber(Config.Checker.Clone(), value).
func (e *Expect) Number(value float64) *Number {
	return NewNumber(e.config.Checker.Clone(), value)
}

// Boolean is a shorthand for NewBoolean(Config.Checker.Clone(), value).
func (e *Expect) Boolean(value bool) *Boolean {
	return NewBoolean(e.config.Checker.Clone(), value)
}
