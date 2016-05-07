// Package httpexpect helps writing nice tests for your HTTP API.
package httpexpect

import (
	"net/http"
	"strings"
	"testing"
)

type Expect struct {
	config Config
}

type Config struct {
	BaseUrl string
	Client  Client
	Checker Checker
	Logger  Logger
}

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Checker interface {
	Clone() Checker
	Compare(a, b interface{}) bool
	Failed() bool
	Fail(message string, args ...interface{})
	Equal(expected, actual interface{})
	NotEqual(expected, actual interface{})
}

type Logger interface {
	LogRequest(method, url string)
}

type DefaultLogger struct {
	*testing.T
}

func (logger DefaultLogger) LogRequest(method, url string) {
	logger.T.Logf("[httpexpect] %s %s", method, url)
}

func New(t *testing.T, baseUrl string) *Expect {
	return WithConfig(Config{
		BaseUrl: baseUrl,
		Checker: NewAssertChecker(t),
		Logger:  DefaultLogger{t},
	})
}

func WithConfig(config Config) *Expect {
	if config.Client == nil {
		config.Client = http.DefaultClient
	}
	if config.Checker == nil {
		panic("config.Checker is nil")
	}
	return &Expect{config}
}

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

func (e *Expect) OPTIONS(url string) *Request {
	return e.Request("OPTIONS", url)
}

func (e *Expect) HEAD(url string) *Request {
	return e.Request("HEAD", url)
}

func (e *Expect) GET(url string) *Request {
	return e.Request("GET", url)
}

func (e *Expect) POST(url string) *Request {
	return e.Request("POST", url)
}

func (e *Expect) PUT(url string) *Request {
	return e.Request("PUT", url)
}

func (e *Expect) PATCH(url string) *Request {
	return e.Request("PATCH", url)
}

func (e *Expect) DELETE(url string) *Request {
	return e.Request("DELETE", url)
}
