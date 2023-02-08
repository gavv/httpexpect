package httpexpect

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpect_Requests(t *testing.T) {
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

func TestExpect_Builders(t *testing.T) {
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

func TestExpect_BuildersCopying(t *testing.T) {
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

func TestExpect_Matchers(t *testing.T) {
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

func TestExpect_MatchersCopying(t *testing.T) {
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

func TestExpect_Values(t *testing.T) {
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

func TestExpect_Traverse(t *testing.T) {
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

	m.IsEqual(data)

	m.ContainsKey("aaa")
	m.ContainsKey("bbb")
	m.ContainsKey("aaa")

	m.IsValueEqual("aaa", data["aaa"])
	m.IsValueEqual("bbb", data["bbb"])
	m.IsValueEqual("ccc", data["ccc"])

	m.Keys().ConsistsOf("aaa", "bbb", "ccc")
	m.Values().ConsistsOf(data["aaa"], data["bbb"], data["ccc"])

	m.Value("aaa").Array().ConsistsOf("bbb", 123, false, nil)
	m.Value("bbb").String().IsEqual("hello")
	m.Value("ccc").Number().IsEqual(456)

	m.Value("aaa").Array().Value(2).Boolean().IsFalse()
	m.Value("aaa").Array().Value(3).IsNull()
}

func TestExpect_Branches(t *testing.T) {
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

	req := WithConfig(config).GET("/url").WithJSON(data)
	resp := req.Expect()

	m1 := resp.JSON().Array()  // fail
	m2 := resp.JSON().Object() // ok
	m3 := resp.JSON().Object() // ok

	e1 := m2.Value("foo").Object()                    // fail
	e2 := m2.Value("foo").Array().Value(999).String() // fail
	e3 := m2.Value("foo").Array().Value(0).Number()   // fail
	e4 := m2.Value("foo").Array().Value(0).String()   // ok
	e5 := m2.Value("foo").Array().Value(0).String()   // ok

	e4.IsEqual("qux") // fail
	e5.IsEqual("bar") // ok

	req.chain.assertFlags(t, flagFailedChildren)
	resp.chain.assertFlags(t, flagFailedChildren)

	m1.chain.assertFlags(t, flagFailed)
	m2.chain.assertFlags(t, flagFailedChildren)
	m3.chain.assertFlags(t, 0)

	e1.chain.assertFlags(t, flagFailed)
	e2.chain.assertFlags(t, flagFailed)
	e3.chain.assertFlags(t, flagFailed)
	e4.chain.assertFlags(t, flagFailed)
	e5.chain.assertFlags(t, 0)
}

func TestExpect_StdCompat(_ *testing.T) {
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

func TestExpect_RequestFactory(t *testing.T) {
	t.Run("default factory", func(t *testing.T) {
		e := WithConfig(Config{
			BaseURL:  "http://example.com",
			Reporter: NewAssertReporter(t),
		})

		req := e.Request("GET", "/")
		req.chain.assertNotFailed(t)

		assert.NotNil(t, req.httpReq)
	})

	t.Run("custom factory", func(t *testing.T) {
		factory := &testRequestFactory{}

		e := WithConfig(Config{
			BaseURL:        "http://example.com",
			Reporter:       NewAssertReporter(t),
			RequestFactory: factory,
		})

		req := e.Request("GET", "/")
		req.chain.assertNotFailed(t)

		assert.NotNil(t, factory.lastreq)
		assert.Same(t, req.httpReq, factory.lastreq)
	})

	t.Run("factory failure", func(t *testing.T) {
		factory := &testRequestFactory{
			fail: true,
		}

		e := WithConfig(Config{
			BaseURL:        "http://example.com",
			Reporter:       newMockReporter(t),
			RequestFactory: factory,
		})

		req := e.Request("GET", "/")
		req.chain.assertFailed(t)

		assert.Nil(t, factory.lastreq)
	})
}

func TestExpect_Panics(t *testing.T) {
	t.Run("nil_AssertionHandler_nonnil_Reporter", func(t *testing.T) {
		assert.NotPanics(t, func() {
			WithConfig(Config{
				Reporter:         newMockReporter(t),
				AssertionHandler: nil,
			})
		})
	})

	t.Run("nonnil_AssertionHandler_nil_Reporter", func(t *testing.T) {
		assert.NotPanics(t, func() {
			WithConfig(Config{
				Reporter:         nil,
				AssertionHandler: &mockAssertionHandler{},
			})
		})
	})

	t.Run("nil_AssertionHandler_nil_Reporter", func(t *testing.T) {
		assert.Panics(t, func() {
			WithConfig(Config{
				Reporter:         nil,
				AssertionHandler: nil,
			})
		})
	})
}
