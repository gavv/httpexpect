package e2e

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
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

	mux.HandleFunc("/echo/", func(w http.ResponseWriter, r *http.Request) {
		arg := strings.TrimPrefix(r.URL.Path, "/echo/")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(arg))
	})

	return mux
}

func testBasicHandler(e *httpexpect.Expect) {
	e.GET("/foo")
	e.GET("/foo").Expect()
	e.GET("/foo").Expect().Status(http.StatusOK)

	e.GET("/foo").Expect().
		Status(http.StatusOK).JSON().Object().HasValue("foo", 123)

	e.PUT("/bar").WithQuery("field1", "hello").WithFormField("field2", "world").
		Expect().
		Status(http.StatusOK).
		Form().HasValue("field1", "hello").HasValue("field2", "world")

	e.GET("/{a}/{b}", "baz", "qux").
		Expect().
		Status(http.StatusOK).JSON().Array().ConsistsOf(true, false)

	e.PUT("/{a}/{b}").
		WithPath("a", "baz").
		WithPath("b", "qux").
		WithJSON(map[string]string{"test": "ok"}).
		Expect().
		Status(http.StatusOK).Body().IsEqual("ok")

	auth := e.Builder(func(req *httpexpect.Request) {
		req.WithBasicAuth("john", "secret")
	})

	auth.PUT("/wee").
		Expect().
		Status(http.StatusOK).
		Form().HasValue("username", "john").HasValue("password", "secret")

	e.PUT("/echo/{arg}").
		WithPath("arg", "test_arg").
		Expect().
		Status(http.StatusOK).Body().IsEqual("test_arg")

	e.PUT("/echo/{arg}").
		WithPath("arg", url.QueryEscape("some:data/thats/encoded@0.1.0")).
		Expect().
		Status(http.StatusOK).
		Body().
		IsEqual(url.QueryEscape("some:data/thats/encoded@0.1.0"))
}

func TestE2EBasic_LiveDefault(t *testing.T) {
	handler := createBasicHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testBasicHandler(httpexpect.Default(t, server.URL))
}

func TestE2EBasic_LiveConfig(t *testing.T) {
	handler := createBasicHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testBasicHandler(httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCurlPrinter(t),
			httpexpect.NewDebugPrinter(t, true),
		},
	}))
}

func TestE2EBasic_LiveTLS(t *testing.T) {
	handler := createBasicHandler()

	server := httptest.NewTLSServer(handler)
	defer server.Close()

	testBasicHandler(httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}))
}

func TestE2EBasic_LiveLongRun(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	handler := createBasicHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	for i := 0; i < 2; i++ {
		testBasicHandler(e)
	}
}

func TestE2EBasic_BinderStandard(t *testing.T) {
	handler := createBasicHandler()

	testBasicHandler(httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
		},
	}))
}

func TestE2EBasic_BinderFast(t *testing.T) {
	handler := fasthttpadaptor.NewFastHTTPHandler(createBasicHandler())

	testBasicHandler(httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
		},
	}))
}
