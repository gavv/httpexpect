package httpexpect

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpectMethods(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	var reqs [8]*Request

	e := WithConfig(config)

	reqs[0] = e.Request("METHOD", "/url")
	reqs[1] = e.OPTIONS("/url")
	reqs[2] = e.HEAD("/url")
	reqs[3] = e.GET("/url")
	reqs[4] = e.POST("/url")
	reqs[5] = e.PUT("/url")
	reqs[6] = e.PATCH("/url")
	reqs[7] = e.DELETE("/url")

	assert.Equal(t, "METHOD", reqs[0].httpReq.Method)
	assert.Equal(t, "OPTIONS", reqs[1].httpReq.Method)
	assert.Equal(t, "HEAD", reqs[2].httpReq.Method)
	assert.Equal(t, "GET", reqs[3].httpReq.Method)
	assert.Equal(t, "POST", reqs[4].httpReq.Method)
	assert.Equal(t, "PUT", reqs[5].httpReq.Method)
	assert.Equal(t, "PATCH", reqs[6].httpReq.Method)
	assert.Equal(t, "DELETE", reqs[7].httpReq.Method)
}

func TestExpectBuilders(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	e := WithConfig(config)

	var reqs1 []*Request

	e1 := e.Builder(func(r *Request) {
		reqs1 = append(reqs1, r)
	})

	var reqs2 []*Request

	e2 := e1.Builder(func(r *Request) {
		reqs2 = append(reqs2, r)
	})

	e.Request("METHOD", "/url")

	r1 := e1.Request("METHOD", "/url")
	r2 := e2.Request("METHOD", "/url")

	assert.Equal(t, 2, len(reqs1))
	assert.Equal(t, 1, len(reqs2))

	assert.Same(t, r1, reqs1[0])
	assert.Same(t, r2, reqs1[1])
	assert.Same(t, r2, reqs2[0])
}

func TestExpectBuildersCopying(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	counter1 := 0
	counter2a := 0
	counter2b := 0

	e0 := WithConfig(config)

	// Simulate the case when many builders are added, and the builders slice
	// have some additioonal capacity. We are going to check that the slice
	// is cloned properly when a new builder is appended.
	for i := 0; i < 10; i++ {
		e0 = e0.Builder(func(r *Request) {})
	}

	e1 := e0.Builder(func(r *Request) {
		counter1++
	})

	e2a := e1.Builder(func(r *Request) {
		counter2a++
	})

	e2b := e1.Builder(func(r *Request) {
		counter2b++
	})

	e0.Request("METHOD", "/url")
	assert.Equal(t, 0, counter1)
	assert.Equal(t, 0, counter2a)
	assert.Equal(t, 0, counter2b)

	e1.Request("METHOD", "/url")
	assert.Equal(t, 1, counter1)
	assert.Equal(t, 0, counter2a)
	assert.Equal(t, 0, counter2b)

	e2a.Request("METHOD", "/url")
	assert.Equal(t, 2, counter1)
	assert.Equal(t, 1, counter2a)
	assert.Equal(t, 0, counter2b)

	e2b.Request("METHOD", "/url")
	assert.Equal(t, 3, counter1)
	assert.Equal(t, 1, counter2a)
	assert.Equal(t, 1, counter2b)
}

func TestExpectMatchers(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	e := WithConfig(config)

	var resps1 []*Response

	e1 := e.Matcher(func(r *Response) {
		resps1 = append(resps1, r)
	})

	var resps2 []*Response

	e2 := e1.Matcher(func(r *Response) {
		resps2 = append(resps2, r)
	})

	e.Request("METHOD", "/url")

	req1 := e1.Request("METHOD", "/url")
	req2 := e2.Request("METHOD", "/url")

	assert.Equal(t, 0, len(resps1))
	assert.Equal(t, 0, len(resps2))

	resp1 := req1.Expect()
	resp2 := req2.Expect()

	assert.Equal(t, 2, len(resps1))
	assert.Equal(t, 1, len(resps2))

	assert.Same(t, resp1, resps1[0])
	assert.Same(t, resp2, resps1[1])
	assert.Same(t, resp2, resps2[0])
}

func TestExpectMatchersCopying(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		Client:   client,
		Reporter: reporter,
	}

	counter1 := 0
	counter2a := 0
	counter2b := 0

	e0 := WithConfig(config)

	// Simulate the case when many builders are added, and the builders slice
	// have some additioonal capacity. We are going to check that the slice
	// is cloned properly when a new builder is appended.
	for i := 0; i < 10; i++ {
		e0 = e0.Matcher(func(r *Response) {})
	}

	e1 := e0.Matcher(func(r *Response) {
		counter1++
	})

	e2a := e1.Matcher(func(r *Response) {
		counter2a++
	})

	e2b := e1.Matcher(func(r *Response) {
		counter2b++
	})

	e0.Request("METHOD", "/url").Expect()
	assert.Equal(t, 0, counter1)
	assert.Equal(t, 0, counter2a)
	assert.Equal(t, 0, counter2b)

	e1.Request("METHOD", "/url").Expect()
	assert.Equal(t, 1, counter1)
	assert.Equal(t, 0, counter2a)
	assert.Equal(t, 0, counter2b)

	e2a.Request("METHOD", "/url").Expect()
	assert.Equal(t, 2, counter1)
	assert.Equal(t, 1, counter2a)
	assert.Equal(t, 0, counter2b)

	e2b.Request("METHOD", "/url").Expect()
	assert.Equal(t, 3, counter1)
	assert.Equal(t, 1, counter2a)
	assert.Equal(t, 1, counter2b)
}

func TestExpectValues(t *testing.T) {
	client := &mockClient{}

	r := NewAssertReporter(t)

	config := Config{
		Client:   client,
		Reporter: r,
	}

	e := WithConfig(config)

	m := map[string]interface{}{}
	a := []interface{}{}
	s := ""
	n := 0.0
	b := false

	assert.NotNil(t, e.Value(m))
	assert.NotNil(t, e.Object(m))
	assert.NotNil(t, e.Array(a))
	assert.NotNil(t, e.String(s))
	assert.NotNil(t, e.Number(n))
	assert.NotNil(t, e.Boolean(b))
}

func TestExpectTraverse(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	data := map[string]interface{}{
		"aaa": []interface{}{"bbb", 123, false, nil},
		"bbb": "hello",
		"ccc": 456,
	}

	resp := WithConfig(config).GET("/url").WithJSON(data).Expect()

	m := resp.JSON().Object()

	m.Equal(data)

	m.ContainsKey("aaa")
	m.ContainsKey("bbb")
	m.ContainsKey("aaa")

	m.ValueEqual("aaa", data["aaa"])
	m.ValueEqual("bbb", data["bbb"])
	m.ValueEqual("ccc", data["ccc"])

	m.Keys().Elements("aaa", "bbb", "ccc")
	m.Values().Elements(data["aaa"], data["bbb"], data["ccc"])

	m.Value("aaa").Array().Elements("bbb", 123, false, nil)
	m.Value("bbb").String().Equal("hello")
	m.Value("ccc").Number().Equal(456)

	m.Value("aaa").Array().Element(2).Boolean().False()
	m.Value("aaa").Array().Element(3).Null()
}

func TestExpectBranches(t *testing.T) {
	client := &mockClient{}

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: newMockReporter(t),
	}

	data := map[string]interface{}{
		"foo": []interface{}{"bar", 123, false, nil},
		"bar": "hello",
		"baz": 456,
	}

	resp := WithConfig(config).GET("/url").WithJSON(data).Expect()

	m1 := resp.JSON().Array()
	m2 := resp.JSON().Object()

	e1 := m2.Value("foo").Object()
	e2 := m2.Value("foo").Array().Element(999).String()
	e3 := m2.Value("foo").Array().Element(0).Number()
	e4 := m2.Value("foo").Array().Element(0).String()

	e4.Equal("bar")

	m1.chain.assertFailed(t)
	m2.chain.assertOK(t)

	e1.chain.assertFailed(t)
	e2.chain.assertFailed(t)
	e3.chain.assertFailed(t)
	e4.chain.assertOK(t)
}

func TestExpectStdCompat(_ *testing.T) {
	Default(&testing.T{}, "")
	Default(&testing.B{}, "")
	Default(testing.TB(&testing.T{}), "")
}

type testRequestFactory struct {
	lastreq *http.Request
	fail    bool
}

func (f *testRequestFactory) NewRequest(
	method, urlStr string, body io.Reader) (*http.Request, error) {
	if f.fail {
		return nil, errors.New("testRequestFactory")
	}
	f.lastreq = httptest.NewRequest(method, urlStr, body)
	return f.lastreq, nil
}

func TestExpectRequestFactory(t *testing.T) {
	e1 := WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
	})
	r1 := e1.Request("GET", "/")
	r1.chain.assertOK(t)
	assert.NotNil(t, r1.httpReq)

	f2 := &testRequestFactory{}
	e2 := WithConfig(Config{
		BaseURL:        "http://example.com",
		Reporter:       NewAssertReporter(t),
		RequestFactory: f2,
	})
	r2 := e2.Request("GET", "/")
	r2.chain.assertOK(t)
	assert.NotNil(t, f2.lastreq)
	assert.True(t, f2.lastreq == r2.httpReq)

	f3 := &testRequestFactory{
		fail: true,
	}
	e3 := WithConfig(Config{
		BaseURL:        "http://example.com",
		Reporter:       newMockReporter(t),
		RequestFactory: f3,
	})
	r3 := e3.Request("GET", "/")
	r3.chain.assertFailed(t)
	assert.Nil(t, f3.lastreq)
}
