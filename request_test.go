package httpexpect

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestRequestFailed(t *testing.T) {
	client := &mockClient{}

	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	config := Config{
		Client: client,
	}

	req := &Request{
		config: config,
		chain:  chain,
	}

	resp := req.Expect()

	assert.False(t, resp == nil)

	req.chain.assertFailed(t)
	resp.chain.assertFailed(t)
}

func TestRequestEmpty(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "", "")

	resp := req.Expect()

	req.chain.assertOK(t)
	resp.chain.assertOK(t)
}

func TestRequestTime(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	for n := 0; n < 10; n++ {
		req := NewRequest(config, "", "")
		resp := req.Expect()
		assert.True(t, resp.time >= 0)
	}
}

func TestRequestURL(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "http://example.com")
	req2 := NewRequest(config, "METHOD", "http://example.com/path")
	req3 := NewRequest(config, "METHOD", "/path")
	req4 := NewRequest(config, "METHOD", "path")

	req1.Expect()
	req1.chain.assertOK(t)
	assert.Equal(t, "http://example.com", client.req.URL.String())

	req2.Expect()
	req2.chain.assertOK(t)
	assert.Equal(t, "http://example.com/path", client.req.URL.String())

	req3.Expect()
	req3.chain.assertOK(t)
	assert.Equal(t, "/path", client.req.URL.String())

	req4.Expect()
	req4.chain.assertOK(t)
	assert.Equal(t, "path", client.req.URL.String())
}

func TestRequestURLQuery(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "http://example.com/path").
		WithQuery("aa", "foo").WithQuery("bb", 123).WithQuery("cc", "*&@")

	q := map[string]interface{}{
		"bb": 123,
		"cc": "*&@",
	}

	req2 := NewRequest(config, "METHOD", "http://example.com/path").
		WithQuery("aa", "foo").
		WithQueryObject(q)

	type S struct {
		Bb int    `url:"bb"`
		Cc string `url:"cc"`
		Dd string `url:"-"`
	}

	req3 := NewRequest(config, "METHOD", "http://example.com/path?aa=foo").
		WithQueryObject(S{123, "*&@", "dummy"})

	req4 := NewRequest(config, "METHOD", "http://example.com/path?aa=foo").
		WithQueryObject(&S{123, "*&@", "dummy"})

	for _, req := range []*Request{req1, req2, req3, req4} {
		client.req = nil

		req.Expect()
		req.chain.assertOK(t)
		assert.Equal(t, "http://example.com/path?aa=foo&bb=123&cc=%2A%26%40",
			client.req.URL.String())
	}

	req5 := NewRequest(config, "METHOD", "http://example.com/path").
		WithQuery("foo", "bar").
		WithQueryObject(nil)

	req5.Expect()
	req5.chain.assertOK(t)
	assert.Equal(t, "http://example.com/path?foo=bar", client.req.URL.String())

	NewRequest(config, "METHOD", "http://example.com/path").
		WithQueryObject(func() {}).chain.assertFailed(t)
}

func TestRequestURLConcat(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	var reqs [5]*Request

	config1 := Config{
		BaseURL:  "",
		Client:   client,
		Reporter: reporter,
	}

	reqs[0] = NewRequest(config1, "METHOD", "http://example.com/path")

	config2 := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	reqs[1] = NewRequest(config2, "METHOD", "path")
	reqs[2] = NewRequest(config2, "METHOD", "/path")

	config3 := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: reporter,
	}

	reqs[3] = NewRequest(config3, "METHOD", "path")
	reqs[4] = NewRequest(config3, "METHOD", "/path")

	for _, req := range reqs {
		assert.Equal(t, "http://example.com/path", req.http.URL.String())
	}

	empty1 := NewRequest(config1, "METHOD", "")
	empty2 := NewRequest(config2, "METHOD", "")
	empty3 := NewRequest(config3, "METHOD", "")

	assert.Equal(t, "", empty1.http.URL.String())
	assert.Equal(t, "http://example.com", empty2.http.URL.String())
	assert.Equal(t, "http://example.com/", empty3.http.URL.String())
}

func TestRequestURLFormat(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	var reqs [3]*Request

	config1 := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: reporter,
	}

	reqs[0] = NewRequest(config1, "METHOD", "/foo/%s", "bar")
	reqs[1] = NewRequest(config1, "METHOD", "%sfoo%s", "/", "/bar")
	reqs[2] = NewRequest(config1, "%s", "/foo/bar")

	for _, req := range reqs {
		assert.Equal(t, "http://example.com/foo/bar", req.http.URL.String())
	}

	config2 := Config{
		Reporter: newMockReporter(t),
	}

	r := NewRequest(config2, "GET", "%s", nil)

	r.chain.assertFailed(t)
}

func TestRequestHeaders(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeader("first-header", "foo")

	req.WithHeaders(map[string]string{
		"Second-Header": "bar",
		"content-Type":  "baz",
		"HOST":          "example.com",
	})

	expectedHeaders := map[string][]string{
		"First-Header":  {"foo"},
		"Second-Header": {"bar"},
		"Content-Type":  {"baz"},
	}

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "example.com", client.req.Host)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestCookies(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithCookie("foo", "1")
	req.WithCookie("bar", "2 ")

	req.WithCookies(map[string]string{
		"baz": " 3",
	})

	expectedHeaders := map[string][]string{
		"Cookie": {`foo=1; bar="2 "; baz=" 3"`},
	}

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyReader(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithChunked(bytes.NewBufferString("body"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.False(t, client.req.Body == nil)
	assert.Equal(t, int64(-1), client.req.ContentLength)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyReaderNil(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithChunked(nil)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.True(t, client.req.Body == nil)
	assert.Equal(t, int64(0), client.req.ContentLength)
}

func TestRequestBodyBytes(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBytes([]byte("body"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.False(t, client.req.Body == nil)
	assert.Equal(t, int64(len("body")), client.req.ContentLength)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, make(http.Header), client.req.Header)
	assert.Equal(t, "body", string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyBytesNil(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithBytes(nil)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.True(t, client.req.Body == nil)
	assert.Equal(t, int64(0), client.req.ContentLength)
}

func TestRequestBodyText(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"text/plain; charset=utf-8"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithText("some text")

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, "some text", string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyForm(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithForm(map[string]interface{}{
		"a": 1,
		"b": "2",
	})

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2`, string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyField(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithFormField("a", 1)
	req.WithFormField("b", "2")

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2`, string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyFormStruct(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	req := NewRequest(config, "METHOD", "url")

	type S struct {
		A string `form:"a"`
		B int    `form:"b"`
		C int    `form:"-"`
	}

	req.WithForm(S{"1", 2, 3})

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2`, string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyFormCombined(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	req := NewRequest(config, "METHOD", "url")

	type S struct {
		A int `form:"a"`
	}

	req.WithForm(S{A: 1})
	req.WithForm(map[string]string{"b": "2"})
	req.WithFormField("c", 3)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `a=1&b=2&c=3`, string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestBodyMultipart(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "POST", "url")

	req.WithMultipart()
	req.WithForm(map[string]string{"b": "1", "c": "2"})
	req.WithFormField("a", 3)

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "POST", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())

	mediatype, params, err := mime.ParseMediaType(client.req.Header.Get("Content-Type"))

	assert.True(t, err == nil)
	assert.Equal(t, "multipart/form-data", mediatype)
	assert.True(t, params["boundary"] != "")

	reader := multipart.NewReader(bytes.NewReader(resp.content), params["boundary"])

	part1, _ := reader.NextPart()
	assert.Equal(t, "b", part1.FormName())
	assert.Equal(t, "", part1.FileName())
	b1, _ := ioutil.ReadAll(part1)
	assert.Equal(t, "1", string(b1))

	part2, _ := reader.NextPart()
	assert.Equal(t, "c", part2.FormName())
	assert.Equal(t, "", part2.FileName())
	b2, _ := ioutil.ReadAll(part2)
	assert.Equal(t, "2", string(b2))

	part3, _ := reader.NextPart()
	assert.Equal(t, "a", part3.FormName())
	assert.Equal(t, "", part3.FileName())
	b3, _ := ioutil.ReadAll(part3)
	assert.Equal(t, "3", string(b3))

	eof, _ := reader.NextPart()
	assert.True(t, eof == nil)
}

func TestRequestBodyMultipartFile(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "POST", "url")

	fh, _ := ioutil.TempFile("", "httpexpect")
	filename2 := fh.Name()
	fh.WriteString("2")
	fh.Close()
	defer os.Remove(filename2)

	req.WithMultipart()
	req.WithForm(map[string]string{"a": "1"})
	req.WithFile("b", filename2)
	req.WithFile("c", "filename3", strings.NewReader("3"))
	req.WithFileBytes("d", "filename4", []byte("4"))

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "POST", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())

	mediatype, params, err := mime.ParseMediaType(client.req.Header.Get("Content-Type"))

	assert.True(t, err == nil)
	assert.Equal(t, "multipart/form-data", mediatype)
	assert.True(t, params["boundary"] != "")

	reader := multipart.NewReader(bytes.NewReader(resp.content), params["boundary"])

	part1, _ := reader.NextPart()
	assert.Equal(t, "a", part1.FormName())
	assert.Equal(t, "", part1.FileName())
	b1, _ := ioutil.ReadAll(part1)
	assert.Equal(t, "1", string(b1))

	part2, _ := reader.NextPart()
	assert.Equal(t, "b", part2.FormName())
	assert.Equal(t, filename2, part2.FileName())
	b2, _ := ioutil.ReadAll(part2)
	assert.Equal(t, "2", string(b2))

	part3, _ := reader.NextPart()
	assert.Equal(t, "c", part3.FormName())
	assert.Equal(t, "filename3", part3.FileName())
	b3, _ := ioutil.ReadAll(part3)
	assert.Equal(t, "3", string(b3))

	part4, _ := reader.NextPart()
	assert.Equal(t, "d", part4.FormName())
	assert.Equal(t, "filename4", part4.FileName())
	b4, _ := ioutil.ReadAll(part4)
	assert.Equal(t, "4", string(b4))

	eof, _ := reader.NextPart()
	assert.True(t, eof == nil)
}

func TestRequestBodyJSON(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	expectedHeaders := map[string][]string{
		"Content-Type": {"application/json; charset=utf-8"},
		"Some-Header":  {"foo"},
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithHeaders(map[string]string{
		"Some-Header": "foo",
	})

	req.WithJSON(map[string]interface{}{"key": "value"})

	resp := req.Expect()
	resp.chain.assertOK(t)

	assert.Equal(t, "METHOD", client.req.Method)
	assert.Equal(t, "url", client.req.URL.String())
	assert.Equal(t, http.Header(expectedHeaders), client.req.Header)
	assert.Equal(t, `{"key":"value"}`, string(resp.content))

	assert.Equal(t, &client.resp, resp.Raw())
}

func TestRequestContentLength(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithChunked(bytes.NewReader([]byte("12345")))
	req1.Expect().chain.assertOK(t)
	assert.Equal(t, int64(-1), client.req.ContentLength)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithBytes([]byte("12345"))
	req2.Expect().chain.assertOK(t)
	assert.Equal(t, int64(5), client.req.ContentLength)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithText("12345")
	req3.Expect().chain.assertOK(t)
	assert.Equal(t, int64(5), client.req.ContentLength)

	j, _ := json.Marshal(map[string]string{"a": "b"})
	req4 := NewRequest(config, "METHOD", "url")
	req4.WithJSON(map[string]string{"a": "b"})
	req4.Expect().chain.assertOK(t)
	assert.Equal(t, int64(len(j)), client.req.ContentLength)

	f := `a=b`
	req5 := NewRequest(config, "METHOD", "url")
	req5.WithForm(map[string]string{"a": "b"})
	req5.Expect().chain.assertOK(t)
	assert.Equal(t, int64(len(f)), client.req.ContentLength)

	req6 := NewRequest(config, "METHOD", "url")
	req6.WithFormField("a", "b")
	req6.Expect().chain.assertOK(t)
	assert.Equal(t, int64(len(f)), client.req.ContentLength)

	req7 := NewRequest(config, "METHOD", "url")
	req7.WithMultipart()
	req7.WithFileBytes("a", "b", []byte("12345"))
	req7.Expect().chain.assertOK(t)
	assert.True(t, client.req.ContentLength > 0)
}

func TestRequestContentTypeOverwrite(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithText("hello")
	req1.WithHeader("Content-Type", "foo")
	req1.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithHeader("Content-Type", "foo")
	req2.WithText("hello")
	req2.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo"}}, client.req.Header)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithJSON(map[string]interface{}{"a": "b"})
	req3.WithHeader("Content-Type", "foo")
	req3.WithHeader("Content-Type", "bar")
	req3.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)

	req4 := NewRequest(config, "METHOD", "url")
	req4.WithForm(map[string]interface{}{"a": "b"})
	req4.WithHeader("Content-Type", "foo")
	req4.WithHeader("Content-Type", "bar")
	req4.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)

	req5 := NewRequest(config, "METHOD", "url")
	req5.WithMultipart()
	req5.WithForm(map[string]interface{}{"a": "b"})
	req5.WithHeader("Content-Type", "foo")
	req5.WithHeader("Content-Type", "bar")
	req5.Expect().chain.assertOK(t)
	assert.Equal(t, http.Header{"Content-Type": {"foo", "bar"}}, client.req.Header)
}

func TestRequestErrorMarshalForm(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithForm(func() {})

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorMarshalJSON(t *testing.T) {
	client := &mockClient{}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithJSON(func() {})

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorReadFile(t *testing.T) {
	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	req.WithMultipart()
	req.WithFile("", "")

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorSend(t *testing.T) {
	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req := NewRequest(config, "METHOD", "url")

	resp := req.Expect()
	resp.chain.assertFailed(t)

	assert.True(t, resp.Raw() == nil)
}

func TestRequestErrorConflictBody(t *testing.T) {
	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithChunked(nil)
	req1.chain.assertOK(t)
	req1.WithChunked(nil)
	req1.chain.assertFailed(t)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithChunked(nil)
	req2.chain.assertOK(t)
	req2.WithBytes(nil)
	req2.chain.assertFailed(t)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithChunked(nil)
	req3.chain.assertOK(t)
	req3.WithText("")
	req3.chain.assertFailed(t)

	req4 := NewRequest(config, "METHOD", "url")
	req4.WithChunked(nil)
	req4.chain.assertOK(t)
	req4.WithJSON(map[string]interface{}{"a": "b"})
	req4.chain.assertFailed(t)

	req5 := NewRequest(config, "METHOD", "url")
	req5.WithChunked(nil)
	req5.chain.assertOK(t)
	req5.WithForm(map[string]interface{}{"a": "b"})
	req5.Expect()
	req5.chain.assertFailed(t)

	req6 := NewRequest(config, "METHOD", "url")
	req6.WithChunked(nil)
	req6.chain.assertOK(t)
	req6.WithFormField("a", "b")
	req6.Expect()
	req6.chain.assertFailed(t)

	req7 := NewRequest(config, "METHOD", "url")
	req7.WithChunked(nil)
	req7.chain.assertOK(t)
	req7.WithMultipart()
	req7.chain.assertFailed(t)
}

func TestRequestErrorConflictType(t *testing.T) {
	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithText("")
	req1.chain.assertOK(t)
	req1.WithJSON(map[string]interface{}{"a": "b"})
	req1.chain.assertFailed(t)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithText("")
	req2.chain.assertOK(t)
	req2.WithForm(map[string]interface{}{"a": "b"})
	req2.chain.assertFailed(t)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithText("")
	req3.chain.assertOK(t)
	req3.WithFormField("a", "b")
	req3.chain.assertFailed(t)

	req4 := NewRequest(config, "METHOD", "url")
	req4.WithText("")
	req4.chain.assertOK(t)
	req4.WithMultipart()
	req4.chain.assertFailed(t)
}

func TestRequestErrorConflictMultipart(t *testing.T) {
	client := &mockClient{
		err: errors.New("error"),
	}

	reporter := newMockReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	req1 := NewRequest(config, "METHOD", "url")
	req1.WithForm(map[string]interface{}{"a": "b"})
	req1.chain.assertOK(t)
	req1.WithMultipart()
	req1.chain.assertFailed(t)

	req2 := NewRequest(config, "METHOD", "url")
	req2.WithFormField("a", "b")
	req2.chain.assertOK(t)
	req2.WithMultipart()
	req2.chain.assertFailed(t)

	req3 := NewRequest(config, "METHOD", "url")
	req3.WithFileBytes("a", "a", []byte("a"))
	req3.chain.assertFailed(t)
}
