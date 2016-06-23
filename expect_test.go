package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/fasthttpexpect"
	"github.com/valyala/fasthttp/fasthttpadaptor"
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

	assert.Equal(t, "METHOD", reqs[0].http.Method)
	assert.Equal(t, "OPTIONS", reqs[1].http.Method)
	assert.Equal(t, "HEAD", reqs[2].http.Method)
	assert.Equal(t, "GET", reqs[3].http.Method)
	assert.Equal(t, "POST", reqs[4].http.Method)
	assert.Equal(t, "PUT", reqs[5].http.Method)
	assert.Equal(t, "PATCH", reqs[6].http.Method)
	assert.Equal(t, "DELETE", reqs[7].http.Method)
}

func TestExpectValue(t *testing.T) {
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

	assert.Equal(t, NewValue(r, m), e.Value(m))
	assert.Equal(t, NewObject(r, m), e.Object(m))
	assert.Equal(t, NewArray(r, a), e.Array(a))
	assert.Equal(t, NewString(r, s), e.String(s))
	assert.Equal(t, NewNumber(r, n), e.Number(n))
	assert.Equal(t, NewBoolean(r, b), e.Boolean(b))
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

func createHandler() http.Handler {
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

	return mux
}

func testHandler(e *Expect) {
	e.GET("/foo")

	e.GET("/foo").Expect()

	e.GET("/foo").Expect().Status(http.StatusOK)

	e.GET("/foo").Expect().
		Status(http.StatusOK).JSON().Object().ValueEqual("foo", 123)

	e.GET("/bar").Expect().
		Status(http.StatusOK).JSON().Array().Elements(true, false)

	e.PUT("/bar").Expect().
		Status(http.StatusNoContent).Body().Empty()
}

func TestExpectLiveDefault(t *testing.T) {
	handler := createHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testHandler(New(t, server.URL))
}

func TestExpectLiveDefaultLongRun(t *testing.T) {
	if testing.Short() {
		return
	}

	handler := createHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := New(t, server.URL)

	for i := 0; i < 2000; i++ {
		testHandler(e)
	}
}

func TestExpectLiveConfig(t *testing.T) {
	handler := createHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testHandler(WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
		Printers: []Printer{
			NewCurlPrinter(t),
			NewDebugPrinter(t, true),
		},
	}))
}

func TestExpectBinderStandard(t *testing.T) {
	handler := createHandler()

	testHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Client:   NewBinder(handler),
		Reporter: NewAssertReporter(t),
	}))
}

func TestExpectBinderFast(t *testing.T) {
	handler := fasthttpadaptor.NewFastHTTPHandler(createHandler())

	testHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Client:   fasthttpexpect.NewBinder(handler),
		Reporter: NewAssertReporter(t),
	}))
}

func createCookieHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "myname",
			Value: "myvalue",
		})
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("myname")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(cookie.Value))
		}
	})

	return mux
}

func testCookies(e *Expect, working bool) {
	e.PUT("/set").Expect().Status(http.StatusNoContent)

	if working {
		e.GET("/get").Expect().Status(http.StatusOK).Text().Equal("myvalue")
	} else {
		e.GET("/get").Expect().Status(http.StatusBadRequest)
	}

}

func TestExpectCookiesClientDisabled(t *testing.T) {
	handler := createCookieHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := WithConfig(Config{
		BaseURL:  server.URL,
		Client:   http.DefaultClient,
		Reporter: NewAssertReporter(t),
	})

	testCookies(e, false)
}

func TestExpectCookiesClientEnabled(t *testing.T) {
	handler := createCookieHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := WithConfig(Config{
		BaseURL:  server.URL,
		Client:   DefaultClient(),
		Reporter: NewAssertReporter(t),
	})

	testCookies(e, true)
}

func TestExpectCookiesBinderStandardDisabled(t *testing.T) {
	handler := createCookieHandler()

	e := WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client:   &Binder{
			Handler: handler,
			Jar:     nil,
		},
	})

	testCookies(e, false)
}

func TestExpectCookiesBinderStandardEnabled(t *testing.T) {
	handler := createCookieHandler()

	e := WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client:   NewBinder(handler),
	})

	testCookies(e, true)
}
