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
//  - non-nil interfaces pointed to nil slices and maps are replaced with
//    nil interfaces
//
// This is equivalent to subsequently aplying json.Marshal() and json.Unmarshal()
// to value.
//
// Failure handling
//
// When some check fails, failure is reported. If non-fatal failures are used
// (see Checker), execution is continued and instance that was checked is marked
// as failed.
//
// If specific instance is marked as failed, all subsequent checks are ignored
// for this instance and any child instances retreived after failure.
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
	// BaseUrl is a URL to prepended to all request. My be empty. If
	// non-empty, trailing slash is allowed but not required and is
	// appended automatically.
	BaseUrl string

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
	// You can use DefaultLogger or provide custom implementation.
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

	// Compare checks if two values are absolutely equal, including their types
	// (binary-equal values of different types are not treated as equal).
	//
	// Compare doesn't report any failures.
	Compare(a, b interface{}) bool

	// Failed checks if some previous check was failed. Clones inherit their
	// failed state from original instance. However, if a clone becomes failed
	// later, it doesn't affect original instance.
	Failed() bool

	// Fail reports failure.
	Fail(message string, args ...interface{})

	// Equal compares two values using Compare and reporst failure if they are
	// not equal.
	//
	// This method exists to allow Checker implementing enhanced failure reporting.
	Equal(expected, actual interface{})

	// NotEqual compares two values using Compare and reporst failure if they are
	// equal.
	//
	// This method exists to allow Checker implementing enhanced failure reporting.
	NotEqual(expected, actual interface{})
}

// Logger is used to report various events.
type Logger interface {
	// LogRequest is called when request is sent.
	LogRequest(method, url string)
}

// DefaultLogger implement Logger. It sends all logs to testing.T instance.
type DefaultLogger struct {
	*testing.T
}

// LogRequest implements Logger.LogRequest.
func (logger DefaultLogger) LogRequest(method, url string) {
	logger.T.Logf("[httpexpect] %s %s", method, url)
}

// New returns a new Expect object.
//
// baseUrl specifies URL to prepended to all request. My be empty. If non-empty,
// trailing slash is allowed but not required and is appended automatically.
//
// New is shorthand for WithConfig. It uses:
//  - http.DefaultClient as Client
//  - AssertChecker as Checker (failures are non-fatal, testify/assert is used)
//  - DefaultLogger as Logger (send logs to testing.T)
//
// Example:
//  func TestAPI(t *testing.T) {
//      e := httpexpect.New(t, "http://example.org/")
//      e.GET("/path").Expect().Status(http.StatusOK)
//  }
func New(t *testing.T, baseUrl string) *Expect {
	return WithConfig(Config{
		BaseUrl: baseUrl,
		Checker: NewAssertChecker(t),
		Logger:  DefaultLogger{t},
	})
}

// WithConfig returns a new Expect object with given config.
//
// If Config.Client is nil, http.DefaultClient is used.
//
// Example:
//  func TestAPI(t *testing.T) {
//      e := httpexpect.WithConfig(httpexpect.Config{
//          BaseUrl: "http://example.org/",
//          Client:  http.DefaultClient,
//          Checker: httpexpect.NewAssertChecker(t),
//          Logger:  httpexpect.DefaultLogger{t},
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
// Config.BaseUrl is prepended to url.
func (e *Expect) Request(method, url string) *Request {
	config := e.config
	config.Checker = config.Checker.Clone()
	return NewRequest(config, method, concatUrls(config.BaseUrl, url))
}

func concatUrls(a, b string) string {
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
		b = b[1:len(b)]
	}
	return a + "/" + b
}

// OPTIONS is a shorthand for Request("OPTIONS", url).
func (e *Expect) OPTIONS(url string) *Request {
	return e.Request("OPTIONS", url)
}

// HEAD is a shorthand for Request("HEAD", url).
func (e *Expect) HEAD(url string) *Request {
	return e.Request("HEAD", url)
}

// GET is a shorthand for Request("GET", url).
func (e *Expect) GET(url string) *Request {
	return e.Request("GET", url)
}

// POST is a shorthand for Request("POST", url).
func (e *Expect) POST(url string) *Request {
	return e.Request("POST", url)
}

// PUT is a shorthand for Request("PUT", url).
func (e *Expect) PUT(url string) *Request {
	return e.Request("PUT", url)
}

// PATCH is a shorthand for Request("PATCH", url).
func (e *Expect) PATCH(url string) *Request {
	return e.Request("PATCH", url)
}

// DELETE is a shorthand for Request("DELETE", url).
func (e *Expect) DELETE(url string) *Request {
	return e.Request("DELETE", url)
}
