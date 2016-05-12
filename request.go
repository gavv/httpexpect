package httpexpect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urllib "net/url"
	"strings"
)

// Request provides methods to incrementally build http.Request object,
// send it, and receive response.
type Request struct {
	config  Config
	chain   chain
	method  string
	url     *urllib.URL
	headers map[string]string
	body    io.Reader
}

// NewRequest returns a new Request object.
//
// method specifies the HTTP method (GET, POST, PUT, etc.).
// url and args are passed to fmt.Sprintf(), with url as format string.
//
// If Config.BaseURL is non-empty, it is prepended to final url,
// separated by slash.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
func NewRequest(config Config, method, url string, args ...interface{}) *Request {
	chain := makeChain(config.Reporter)

	for _, a := range args {
		if a == nil {
			chain.fail(
				"\nunexpected nil argument for url format string:\n"+
					"  Request(\"%s\", %v...)", method, args)
		}
	}

	urlStr := concatURLs(config.BaseURL, fmt.Sprintf(url, args...))

	urlObj, err := urllib.Parse(urlStr)
	if err != nil {
		chain.fail(err.Error())
	}

	return &Request{
		config: config,
		chain:  chain,
		method: method,
		url:    urlObj,
	}
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

// WithQuery adds query parameter to request URL.
//
// value is converted to string using fmt.Sprint() and urlencoded.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithQuery("foo", 123)
//  req.WithQuery("bar", "baz")
//  // URL is now http://example.org/path?foo=123&bar=baz
func (r *Request) WithQuery(key string, value interface{}) *Request {
	q := r.url.Query()
	q.Add(key, fmt.Sprint(value))
	r.url.RawQuery = q.Encode()
	return r
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
		r.chain.fail(err.Error())
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
	return &Response{
		chain: r.chain,
		resp:  resp,
	}
}

func (r *Request) sendRequest() *http.Response {
	if r.chain.failed() {
		return nil
	}

	req, err := http.NewRequest(r.method, r.url.String(), r.body)
	if err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	if r.config.Printer != nil {
		r.config.Printer.Request(req)
	}

	resp, err := r.config.Client.Do(req)
	if err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	if r.config.Printer != nil {
		r.config.Printer.Response(resp)
	}

	return resp
}
