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
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Request provides methods to incrementally build http.Request object,
// send it, and receive response.
type Request struct {
	config     Config
	chain      chain
	http       http.Request
	query      url.Values
	form       url.Values
	multipart  *multipart.Writer
	forcetype  bool
	typesetter string
	bodysetter string
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
	switch strings.ToLower(k) {
	case "host":
		r.http.Host = v
	case "content-type":
		if !r.forcetype {
			delete(r.http.Header, "Content-Type")
		}
		r.forcetype = true
		r.typesetter = "WithHeader"
		r.http.Header.Add(k, v)
	default:
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
	r.setBody("WithBody", reader, -1)
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
		r.setBody("WithBytes", nil, 0)
	} else {
		r.setBody("WithBytes", bytes.NewReader(b), len(b))
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
	r.setType("WithText", "text/plain; charset=utf-8", false)
	r.setBody("WithText", strings.NewReader(s), -1)
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

	r.setType("WithJSON", "application/json; charset=utf-8", false)
	r.setBody("WithJSON", bytes.NewReader(b), len(b))

	return r
}

// WithForm sets Content-Type header to "application/x-www-form-urlencoded"
// or (if WithMultipart() was called) "multipart/form-data", converts given
// object to url.Values using github.com/ajg/form and adds it to request body.
//
// Various object types are supported, including maps and structs. Structs may
// contain "form" struct tag, similar to "json" struct tag for json.Marshal().
// See https://github.com/ajg/form for details.
//
// Multiple WithForm(), WithField(), and WithFile() calls may be combined.
// If WithMultipart() is called, it should be called first.
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

	if r.multipart != nil {
		r.setType("WithForm", "multipart/form-data", false)

		var keys []string
		for k := range f {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if err := r.multipart.WriteField(k, f[k][0]); err != nil {
				r.chain.fail(err.Error())
				return r
			}
		}
	} else {
		r.setType("WithForm", "application/x-www-form-urlencoded", false)

		if r.form == nil {
			r.form = make(url.Values)
		}
		for k, v := range f {
			r.form[k] = append(r.form[k], v...)
		}
	}

	return r
}

// WithField sets Content-Type header to "application/x-www-form-urlencoded"
// or (if WithMultipart() was called) "multipart/form-data", converts given
// value to string using fmt.Sprint() and adds it to request body.
//
// Multiple WithForm(), WithField(), and WithFile() calls may be combined.
// If WithMultipart() is called, it should be called first.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithField("foo", 123)
func (r *Request) WithField(key string, value interface{}) *Request {
	if r.multipart != nil {
		r.setType("WithField", "multipart/form-data", false)

		err := r.multipart.WriteField(key, fmt.Sprint(value))
		if err != nil {
			r.chain.fail(err.Error())
			return r
		}
	} else {
		r.setType("WithField", "application/x-www-form-urlencoded", false)

		if r.form == nil {
			r.form = make(url.Values)
		}
		r.form[key] = append(r.form[key], fmt.Sprint(value))
	}
	return r
}

// WithFile sets Content-Type header to "multipart/form-data", reads given
// file and adds its contents to request body.
//
// If reader is given, it's used to read file contents. Otherwise, os.Open()
// is used to read a file with given path.
//
// Multiple WithForm(), WithField(), and WithFile() calls may be combined.
// WithMultipart() should be called before WithFile(), otherwise WithFile()
// fails.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithFile("avatar", "./john.png")
//
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  fh, _ := os.Open("./john.png")
//  req.WithMultipart().
//      WithFile("avatar", "john.png", fh)
//  fh.Close()
func (r *Request) WithFile(key, path string, reader ...io.Reader) *Request {
	r.setType("WithFile", "multipart/form-data", false)

	if r.multipart == nil {
		r.chain.fail("WithFile requires WithMultipart to be called first")
		return r
	}

	wr, err := r.multipart.CreateFormFile(key, path)
	if err != nil {
		r.chain.fail(err.Error())
		return r
	}

	var rd io.Reader
	if len(reader) != 0 && reader[0] != nil {
		rd = reader[0]
	} else {
		f, err := os.Open(path)
		if err != nil {
			r.chain.fail(err.Error())
			return r
		}
		rd = f
		defer f.Close()
	}

	if _, err := io.Copy(wr, rd); err != nil {
		r.chain.fail(err.Error())
		return r
	}

	return r
}

// WithFileBytes is like WithFile, but uses given slice of bytes as the
// file contents.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  fh, _ := os.Open("./john.png")
//  b, _ := ioutil.ReadAll(fh)
//  req.WithMultipart().
//      WithFileBytes("avatar", "john.png", b)
//  fh.Close()
func (r *Request) WithFileBytes(key, path string, data []byte) *Request {
	return r.WithFile(key, path, bytes.NewReader(data))
}

// WithMultipart sets Content-Type header to "multipart/form-data".
//
// After this call, WithForm() and WithField() switch to multipart form
// instead of urlencoded form.
//
// If WithMultipart() is called, it should be called before WithForm()
// or WithField().
//
// WithFile() always requires WithMultipart() to be called first.
//
// Example:
//  req := NewRequest(config, "PUT", "http://example.org/path")
//  req.WithMultipart().
//      WithForm(map[string]interface{}{"foo": 123})
func (r *Request) WithMultipart() *Request {
	r.setType("WithMultipart", "multipart/form-data", false)

	if r.multipart == nil {
		var buf bytes.Buffer
		r.multipart = multipart.NewWriter(&buf)
		r.setBody("WithMultipart", &buf, -1)
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
	r.encodeRequest()

	resp, elapsed := r.sendRequest()

	return makeResponse(r.chain, resp, elapsed)
}

func (r *Request) setType(newSetter, newType string, overwrite bool) {
	if r.forcetype {
		return
	}

	if !overwrite {
		previousType := r.http.Header.Get("Content-Type")

		if previousType != "" && previousType != newType {
			r.chain.fail(
				"\nambiguous request \"Content-Type\" header values:\n  %s (set by %s)\n\n"+
					"and:\n  %s (wanted by %s)",
				strconv.Quote(previousType), r.typesetter,
				strconv.Quote(newType), newSetter)
			return
		}
	}

	r.typesetter = newSetter
	r.http.Header["Content-Type"] = []string{newType}
}

func (r *Request) setBody(setter string, reader io.Reader, len int) {
	if r.bodysetter != "" {
		r.chain.fail(
			"\nambiguous request body contents:\n  set by %s\n  overwritten by %s",
			r.bodysetter, setter)
		return
	}

	if reader == nil {
		r.http.Body = nil
		r.http.ContentLength = 0
	} else {
		r.http.Body = ioutil.NopCloser(reader)
		r.http.ContentLength = int64(len)
	}

	r.bodysetter = setter
}

func (r *Request) encodeRequest() {
	if r.query != nil {
		r.http.URL.RawQuery = r.query.Encode()
	}

	if r.multipart != nil {
		r.setType("Expect", r.multipart.FormDataContentType(), true)

		if err := r.multipart.Close(); err != nil {
			r.chain.fail(err.Error())
			return
		}
	} else if r.form != nil {
		r.setBody("WithForm or WithField",
			strings.NewReader(r.form.Encode()), -1)
	}
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

	return
}
