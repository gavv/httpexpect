package httpexpect

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpect_Constructors(t *testing.T) {
	t.Run("testing.T", func(t *testing.T) {
		_ = Default(&testing.T{}, "")
	})

	t.Run("testing.B", func(t *testing.T) {
		_ = Default(&testing.B{}, "")
	})

	t.Run("testing.TB", func(t *testing.T) {
		_ = Default(testing.TB(&testing.T{}), "")
	})
}

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

	reqs[0] = e.Request("GET", "/url")
	reqs[1] = e.OPTIONS("/url")
	reqs[2] = e.HEAD("/url")
	reqs[3] = e.GET("/url")
	reqs[4] = e.POST("/url")
	reqs[5] = e.PUT("/url")
	reqs[6] = e.PATCH("/url")
	reqs[7] = e.DELETE("/url")

	assert.Equal(t, "GET", reqs[0].httpReq.Method)
	assert.Equal(t, "OPTIONS", reqs[1].httpReq.Method)
	assert.Equal(t, "HEAD", reqs[2].httpReq.Method)
	assert.Equal(t, "GET", reqs[3].httpReq.Method)
	assert.Equal(t, "POST", reqs[4].httpReq.Method)
	assert.Equal(t, "PUT", reqs[5].httpReq.Method)
	assert.Equal(t, "PATCH", reqs[6].httpReq.Method)
	assert.Equal(t, "DELETE", reqs[7].httpReq.Method)
}

func TestExpect_Builders(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
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

		e.Request("GET", "/url")

		r1 := e1.Request("GET", "/url")
		r2 := e2.Request("GET", "/url")

		assert.Equal(t, 2, len(reqs1))
		assert.Equal(t, 1, len(reqs2))

		assert.Same(t, r1, reqs1[0])
		assert.Same(t, r2, reqs1[1])
		assert.Same(t, r2, reqs2[0])
	})

	t.Run("copying", func(t *testing.T) {
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

		e0.Request("GET", "/url")
		assert.Equal(t, 0, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e1.Request("GET", "/url")
		assert.Equal(t, 1, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e2a.Request("GET", "/url")
		assert.Equal(t, 2, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 0, counter2b)

		e2b.Request("GET", "/url")
		assert.Equal(t, 3, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 1, counter2b)
	})
}

func TestExpect_Matchers(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
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

		e.Request("GET", "/url")

		req1 := e1.Request("GET", "/url")
		req2 := e2.Request("GET", "/url")

		assert.Equal(t, 0, len(resps1))
		assert.Equal(t, 0, len(resps2))

		resp1 := req1.Expect()
		resp2 := req2.Expect()

		assert.Equal(t, 2, len(resps1))
		assert.Equal(t, 1, len(resps2))

		assert.Same(t, resp1, resps1[0])
		assert.Same(t, resp2, resps1[1])
		assert.Same(t, resp2, resps2[0])
	})

	t.Run("copying", func(t *testing.T) {
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

		e0.Request("GET", "/url").Expect()
		assert.Equal(t, 0, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e1.Request("GET", "/url").Expect()
		assert.Equal(t, 1, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e2a.Request("GET", "/url").Expect()
		assert.Equal(t, 2, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 0, counter2b)

		e2b.Request("GET", "/url").Expect()
		assert.Equal(t, 3, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 1, counter2b)
	})
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
		"foo": []interface{}{"bar", 123, false, nil},
		"bar": "hello",
		"baz": 456,
	}

	resp := WithConfig(config).GET("/url").WithJSON(data).Expect()
	m := resp.JSON().Object()
	m.Equal(data)
	m.ContainsKey("foo")
	m.ContainsKey("bar")
	m.ContainsKey("foo")
	m.ValueEqual("foo", data["foo"])
	m.ValueEqual("bar", data["bar"])
	m.ValueEqual("baz", data["baz"])
	m.Keys().ContainsOnly("foo", "bar", "baz")
	m.Values().ContainsOnly(data["foo"], data["bar"], data["baz"])
	m.Value("foo").Array().Elements("bar", 123, false, nil)
	m.Value("bar").String().Equal("hello")
	m.Value("baz").Number().Equal(456)
	m.Value("foo").Array().Element(2).Boolean().False()
	m.Value("foo").Array().Element(3).Null()
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

	m1 := resp.JSON().Array()
	m2 := resp.JSON().Object()
	m3 := resp.JSON().Object()

	e1 := m2.Value("foo").Object()
	e2 := m2.Value("foo").Array().Value(999).String()
	e3 := m2.Value("foo").Array().Value(0).Number()
	e4 := m2.Value("foo").Array().Value(0).String()
	e5 := m2.Value("foo").Array().Value(0).String()

	e4.IsEqual("qux")
	e5.IsEqual("bar")

	req.chain.assertFlags(t, flagFailedChildren)
	resp.chain.assertFlags(t, flagFailedChildren)

	m1.chain.assertFlags(t, flagFailed)
	m2.chain.assertFlags(t, flagFailedChildren)
	m3.chain.assertFlags(t, 0)

	e1.chain.assertFlags(t, flagFailed)
	e2.chain.assertFlags(t, flagFailed)
	fmt.Println(e3.chain.flags)
	e3.chain.assertFlags(t, flagFailed)
	e4.chain.assertFlags(t, flagFailed)
	e5.chain.assertFlags(t, 0)
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
		factory := &mockRequestFactory{}

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
		factory := &mockRequestFactory{
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
	t.Run("nil AssertionHandler, non-nil Reporter", func(t *testing.T) {
		assert.NotPanics(t, func() {
			WithConfig(Config{
				Reporter:         newMockReporter(t),
				AssertionHandler: nil,
			})
		})
	})

	t.Run("non-nil AssertionHandler, nil Reporter", func(t *testing.T) {
		assert.NotPanics(t, func() {
			WithConfig(Config{
				Reporter:         nil,
				AssertionHandler: &mockAssertionHandler{},
			})
		})
	})

	t.Run("nil AssertionHandler, nil Reporter", func(t *testing.T) {
		assert.Panics(t, func() {
			WithConfig(Config{
				Reporter:         nil,
				AssertionHandler: nil,
			})
		})
	})
}

func TestExpect_Config(t *testing.T) {
	t.Run("defaults, non-nil Reporter", func(t *testing.T) {
		config := Config{
			Reporter: newMockReporter(t),
		}

		config = config.withDefaults()

		assert.NotNil(t, config.RequestFactory)
		assert.NotNil(t, config.Client)
		assert.NotNil(t, config.WebsocketDialer)
		assert.NotNil(t, config.AssertionHandler)
		assert.NotNil(t, config.Formatter)
		assert.NotNil(t, config.Reporter)

		assert.NotPanics(t, func() {
			config.validate()
		})
	})

	t.Run("defaults, non-nil AssertionHandler", func(t *testing.T) {
		config := Config{
			AssertionHandler: &mockAssertionHandler{},
		}

		config = config.withDefaults()

		assert.NotNil(t, config.RequestFactory)
		assert.NotNil(t, config.Client)
		assert.NotNil(t, config.WebsocketDialer)
		assert.NotNil(t, config.AssertionHandler)
		assert.Nil(t, config.Formatter)
		assert.Nil(t, config.Reporter)

		assert.NotPanics(t, func() {
			config.validate()
		})
	})

	t.Run("defaults, nil Reporter and AssertionHandler", func(t *testing.T) {
		config := Config{}

		assert.Panics(t, func() {
			config.withDefaults()
		})

		assert.Panics(t, func() {
			config.validate()
		})
	})

	t.Run("validate fields", func(t *testing.T) {
		config := Config{
			Reporter: newMockReporter(t),
		}

		config = config.withDefaults()

		assert.NotPanics(t, func() {
			config.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.RequestFactory = nil
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.Client = nil
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = nil
			badConfig.validate()
		})
	})

	t.Run("validate handler", func(t *testing.T) {
		config := Config{
			Reporter: newMockReporter(t),
		}

		config = config.withDefaults()

		assert.NotPanics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = &DefaultAssertionHandler{
				Formatter: &DefaultFormatter{},
				Reporter:  newMockReporter(t),
			}
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = &DefaultAssertionHandler{
				Formatter: &DefaultFormatter{},
				Reporter:  nil,
			}
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = &DefaultAssertionHandler{
				Formatter: nil,
				Reporter:  newMockReporter(t),
			}
			badConfig.validate()
		})
	})
}
