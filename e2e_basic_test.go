package httpexpect

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func createBasicHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/foo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	mux.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = w.Write([]byte(`field1=` + r.FormValue("field1")))
		_, _ = w.Write([]byte(`&field2=` + r.PostFormValue("field2")))
	})

	mux.HandleFunc("/baz/qux", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[true, false]`))

		case "PUT":
			decoder := json.NewDecoder(r.Body)
			var m map[string]interface{}
			if err := decoder.Decode(&m); err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else if m["test"] != "ok" {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`ok`))
			}
		}
	})

	mux.HandleFunc("/wee", func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); ok {
			w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
			_, _ = w.Write([]byte(`username=` + u))
			_, _ = w.Write([]byte(`&password=` + p))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	return mux
}

func testBasicHandler(e *Expect) {
	e.GET("/foo")
	e.GET("/foo").Expect()
	e.GET("/foo").Expect().Status(http.StatusOK)

	e.GET("/foo").Expect().
		Status(http.StatusOK).JSON().Object().ValueEqual("foo", 123)

	e.PUT("/bar").WithQuery("field1", "hello").WithFormField("field2", "world").
		Expect().
		Status(http.StatusOK).
		Form().ValueEqual("field1", "hello").ValueEqual("field2", "world")

	e.GET("/{a}/{b}", "baz", "qux").
		Expect().
		Status(http.StatusOK).JSON().Array().Elements(true, false)

	e.PUT("/{a}/{b}").
		WithPath("a", "baz").
		WithPath("b", "qux").
		WithJSON(map[string]string{"test": "ok"}).
		Expect().
		Status(http.StatusOK).Body().Equal("ok")

	auth := e.Builder(func(req *Request) {
		req.WithBasicAuth("john", "secret")
	})

	auth.PUT("/wee").
		Expect().
		Status(http.StatusOK).
		Form().ValueEqual("username", "john").ValueEqual("password", "secret")
}

func TestE2EBasicLiveDefault(t *testing.T) {
	handler := createBasicHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testBasicHandler(New(t, server.URL))
}

func TestE2EBasicLiveConfig(t *testing.T) {
	handler := createBasicHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testBasicHandler(WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
		Printers: []Printer{
			NewCurlPrinter(t),
			NewDebugPrinter(t, true),
		},
	}))
}

func TestE2EBasicLiveTLS(t *testing.T) {
	handler := createBasicHandler()

	server := httptest.NewTLSServer(handler)
	defer server.Close()

	testBasicHandler(WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}))
}

func TestE2EBasicLiveLongRun(t *testing.T) {
	if testing.Short() {
		return
	}

	handler := createBasicHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := New(t, server.URL)

	for i := 0; i < 2; i++ {
		testBasicHandler(e)
	}
}

func TestE2EBasicBinderStandard(t *testing.T) {
	handler := createBasicHandler()

	testBasicHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewBinder(handler),
		},
	}))
}

func TestE2EBasicBinderFast(t *testing.T) {
	handler := fasthttpadaptor.NewFastHTTPHandler(createBasicHandler())

	testBasicHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewFastBinder(handler),
		},
	}))
}
