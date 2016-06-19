package httpexpect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ajg/form"
	"github.com/gavv/monotime"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Request provides methods to incrementally build http.Request object,
// send it, and receive response.
type Request struct {
	config Config
	chain  chain
	http   http.Request
	query  url.Values
	form   url.Values
}

// NewRequest returns a new Request object.
//
// method specifies the HTTP method (GET, POST, PUT, etc.).
// urlfmt and args are passed to fmt.Sprintf(), with url as format string.
//
// If Config.BaseURL is non-empty, it is prepended to final url,
// separated by slash.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
func NewRequest(config Config, method, urlfmt string, args ...interface{}) *Request {
	chain := makeChain(config.Reporter)

	for _, a := range args {
		if a == nil {
			chain.fail(
				"\nunexpected nil argument for url format string:\n"+
					"  Request(\"%s\", %v...)", method, args)
		}
	}

	us := concatURLs(config.BaseURL, fmt.Sprintf(urlfmt, args...))

	u, err := url.Parse(us)
	if err != nil {
		chain.fail(err.Error())
	}

	req := Request{
		config: config,
		chain:  chain,
		http: http.Request{
			Method: method,
			URL:    u,
			Header: make(http.Header),
		},
	}

	return &req
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
//  req.WithQuery("a", 123)
//  req.WithQuery("b", "foo")
//  // URL is now http://example.org/path?a=123&b=foo
func (r *Request) WithQuery(key string, value interface{}) *Request {
	if r.query == nil {
		r.query = r.http.URL.Query()
	}
	r.query.Add(key, fmt.Sprint(value))
	return r
}

// WithQueryObject adds multiple query parameters to request URL.
//
// object is converted to query string using github.com/google/go-querystring
// if it's a struct or pointer to struct, or github.com/ajg/form otherwise.
//
// Various object types are supported. Structs may contain "url" struct tag,
// similar to "json" struct tag for json.Marshal().
//
// Example:
//  type MyURL struct {
//      A int    `url:"a"`
//      B string `url:"b"`
//  }
//
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithQueryObject(MyURL{A: 123, B: "foo"})
//  // URL is now http://example.org/path?a=123&b=foo
//
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithQueryObject(map[string]interface{}{"a": 123, "b": "foo"})
//  // URL is now http://example.org/path?a=123&b=foo
func (r *Request) WithQueryObject(object interface{}) *Request {
	if object == nil {
		return r
	}
	var (
		q   url.Values
		err error
	)
	if reflect.Indirect(reflect.ValueOf(object)).Kind() == reflect.Struct {
		q, err = query.Values(object)
		if err != nil {
			r.chain.fail(err.Error())
			return r
		}
	} else {
		q, err = form.EncodeToValues(object)
		if err != nil {
			r.chain.fail(err.Error())
			return r
		}
	}
	if r.query == nil {
		r.query = r.http.URL.Query()
	}
	for k, v := range q {
		r.query[k] = append(r.query[k], v...)
	}
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
	for k, v := range headers {
		r.WithHeader(k, v)
	}
	return r
}

// WithHeader adds given single header to request.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithHeader("Content-Type": "application/json")
func (r *Request) WithHeader(k, v string) *Request {
	if k == "Host" {
		r.http.Host = v
	} else {
		r.http.Header.Add(k, v)
	}
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
	if reader == nil {
		r.http.Body = nil
		r.http.ContentLength = 0
	} else {
		r.http.Body = ioutil.NopCloser(reader)
		r.http.ContentLength = -1
	}
	return r
}

// WithBytes is like WithBody, but gets body as a slice of bytes.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithHeader("Content-Type": "application/json")
//  req.WithBytes([]byte(`{"foo": 123}`))
func (r *Request) WithBytes(b []byte) *Request {
	if b == nil {
		r.http.Body = nil
		r.http.ContentLength = 0
	} else {
		r.http.Body = ioutil.NopCloser(bytes.NewReader(b))
		r.http.ContentLength = int64(len(b))
	}
	return r
}

// WithText sets Content-Type header to "text/plain; charset=utf-8" and
// sets body to given string.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithText("hello, world!")
func (r *Request) WithText(s string) *Request {
	r.WithHeader("Content-Type", "text/plain; charset=utf-8")
	r.WithBody(strings.NewReader(s))

	return r
}

// WithJSON sets Content-Type header to "application/json; charset=utf-8"
// and sets body to object, marshaled using json.Marshal().
//
// Example:
//  type MyJSON struct {
//      Foo int `json:"foo"`
//  }
//
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithJSON(MyJSON{Foo: 123})
//
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

// WithForm sets Content-Type header to "application/x-www-form-urlencoded"
// and sets body to object, marshaled using github.com/ajg/form module.
//
// Various object types are supported, including maps and structs. Structs may
// contain "form" struct tag, similar to "json" struct tag for json.Marshal().
// See https://github.com/ajg/form for details.
//
// Example:
//  type MyForm struct {
//      Foo int `form:"foo"`
//  }
//
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithForm(MyForm{Foo: 123})
//
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithForm(map[string]interface{}{"foo": 123})
func (r *Request) WithForm(object interface{}) *Request {
	f, err := form.EncodeToValues(object)
	if err != nil {
		r.chain.fail(err.Error())
		return r
	}
	if r.form == nil {
		r.form = make(url.Values)
	}
	for k, v := range f {
		r.form[k] = append(r.form[k], v...)
	}
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
	if r.query != nil {
		r.http.URL.RawQuery = r.query.Encode()
	}

	if r.form != nil {
		r.WithHeader("Content-Type", "application/x-www-form-urlencoded")
		r.WithBody(strings.NewReader(r.form.Encode()))
	}

	resp, elapsed := r.sendRequest()

	return makeResponse(r.chain, resp, elapsed)
}

func (r *Request) sendRequest() (resp *http.Response, elapsed time.Duration) {
	if r.chain.failed() {
		return
	}

	for _, printer := range r.config.Printers {
		printer.Request(&r.http)
	}

	start := monotime.Now()

	resp, err := r.config.Client.Do(&r.http)

	elapsed = monotime.Since(start)

	if err != nil {
		r.chain.fail(err.Error())
		return
	}

	for _, printer := range r.config.Printers {
		printer.Response(resp, elapsed)
	}

	return resp, elapsed
}
