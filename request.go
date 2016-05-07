package httpexpect

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Request provides methods to incrementally build http.Request object,
// send it, and receive response.
type Request struct {
	config  Config
	method  string
	url     string
	headers map[string]string
	body    io.Reader
}

// NewRequest returns a new Request object.
//
// method specifies the HTTP method (GET, POST, PUT, etc.).
// url specifies absolute URL to access.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
func NewRequest(config Config, method, url string) *Request {
	return &Request{
		config: config,
		method: method,
		url:    url,
	}
}

// WithHeaders adds given headers to request.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithHeaders(map[string]string{
//      "Content-Type": "application/json",
//  })
func (r *Request) WithHeaders(headers map[string]string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	for k, v := range headers {
		r.headers[k] = v
	}
	return r
}

// WithHeader adds given single header to request.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithHeader("Content-Type": "application/json")
func (r *Request) WithHeader(k, v string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[k] = v
	return r
}

// WithBody set given reader for request body.
//
// Expect() will read all available data from this reader.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithHeader("Content-Type": "application/json")
//  req.WithBody(bytes.NewBufferString(`{"foo": 123}`))
func (r *Request) WithBody(reader io.Reader) *Request {
	r.body = reader
	return r
}

// WithBytes is like WithBody, but gets body as a slice of bytes.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithHeader("Content-Type": "application/json")
//  req.WithBytes([]byte(`{"foo": 123}`))
func (r *Request) WithBytes(b []byte) *Request {
	return r.WithBody(bytes.NewReader(b))
}

// WithJSON sets Content-Type header to "application/json" and sets body to
// marshaled object.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithJSON(map[string]interface{}{"foo": 123})
func (r *Request) WithJSON(object interface{}) *Request {
	b, err := json.Marshal(object)
	if err != nil {
		r.config.Checker.Fail(err.Error())
		return r
	}

	r.WithHeader("Content-Type", "application/json; charset=utf-8")
	r.WithBytes(b)

	return r
}

// Expect constructs http.Request, sends it, receives http.Response, and
// returns a new Response object to inspect received response.
//
// Request is sent using Config.Client interface.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithJSON(map[string]interface{}{"foo": 123})
//  resp := req.Expect()
//  resp.Status(http.StatusOK)
func (r *Request) Expect() *Response {
	resp := r.sendRequest()
	return NewResponse(r.config.Checker.Clone(), resp)
}

func (r *Request) sendRequest() *http.Response {
	if r.config.Checker.Failed() {
		return nil
	}

	if r.config.Logger != nil {
		r.config.Logger.LogRequest(r.method, r.url)
	}

	req, err := http.NewRequest(r.method, r.url, r.body)
	if err != nil {
		r.config.Checker.Fail(err.Error())
		return nil
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	resp, err := r.config.Client.Do(req)
	if err != nil {
		r.config.Checker.Fail(err.Error())
		return nil
	}

	return resp
}
