package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExpectMethods(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
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

	assert.Equal(t, "METHOD", reqs[0].method)
	assert.Equal(t, "OPTIONS", reqs[1].method)
	assert.Equal(t, "HEAD", reqs[2].method)
	assert.Equal(t, "GET", reqs[3].method)
	assert.Equal(t, "POST", reqs[4].method)
	assert.Equal(t, "PUT", reqs[5].method)
	assert.Equal(t, "PATCH", reqs[6].method)
	assert.Equal(t, "DELETE", reqs[7].method)
}

func TestExpectURLConcat(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	var reqs [5]*Request

	config1 := Config{
		BaseURL:  "",
		Client:   client,
		Reporter: reporter,
	}

	reqs[0] = WithConfig(config1).Request("METHOD", "http://example.com/path")

	config2 := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	reqs[1] = WithConfig(config2).Request("METHOD", "path")
	reqs[2] = WithConfig(config2).Request("METHOD", "/path")

	config3 := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: reporter,
	}

	reqs[3] = WithConfig(config3).Request("METHOD", "path")
	reqs[4] = WithConfig(config3).Request("METHOD", "/path")

	for _, req := range reqs {
		assert.Equal(t, "http://example.com/path", req.url.String())
	}

	empty1 := WithConfig(config1).Request("METHOD", "")
	empty2 := WithConfig(config2).Request("METHOD", "")
	empty3 := WithConfig(config3).Request("METHOD", "")

	assert.Equal(t, "", empty1.url.String())
	assert.Equal(t, "http://example.com", empty2.url.String())
	assert.Equal(t, "http://example.com/", empty3.url.String())
}

func TestExpectURLFormat(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	var reqs [9]*Request

	config := Config{
		BaseURL:  "http://example.com/",
		Client:   client,
		Reporter: reporter,
	}

	reqs[0] = WithConfig(config).Request("METHOD", "/foo/%s", "bar")
	reqs[1] = WithConfig(config).Request("METHOD", "%sfoo%s", "/", "/bar")
	reqs[2] = WithConfig(config).OPTIONS("%s", "/foo/bar")
	reqs[3] = WithConfig(config).HEAD("%s", "/foo/bar")
	reqs[4] = WithConfig(config).GET("%s", "/foo/bar")
	reqs[5] = WithConfig(config).POST("%s", "/foo/bar")
	reqs[6] = WithConfig(config).PUT("%s", "/foo/bar")
	reqs[7] = WithConfig(config).PATCH("%s", "/foo/bar")
	reqs[8] = WithConfig(config).DELETE("%s", "/foo/bar")

	for _, req := range reqs {
		assert.Equal(t, "http://example.com/foo/bar", req.url.String())
	}

	e := WithConfig(Config{
		Reporter: mockReporter{t},
	})

	r := e.Request("GET", "%s", nil)

	r.chain.assertFailed(t)
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

func TestExpectBranches(t *testing.T) {
	client := &mockClient{}

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: mockReporter{t},
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

func TestExpectLive(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/foo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"foo":123}`))
	})

	mux.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[true, false]`))

		case "PUT":
			w.WriteHeader(http.StatusNoContent)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	e := New(t, server.URL)

	e.GET("/foo").Expect().
		Status(http.StatusOK).JSON().Object().ValueEqual("foo", 123)

	e.GET("/bar").Expect().
		Status(http.StatusOK).JSON().Array().Elements(true, false)

	e.PUT("/bar").Expect().
		Status(http.StatusNoContent).NoContent()
}
