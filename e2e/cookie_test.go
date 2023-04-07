package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func createCookieHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "myname",
			Value:   "myvalue",
			Path:    "/",
			Expires: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
		})
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("myname")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(cookie.Value))
		}
	})

	return mux
}

func testCookieHandler(e *httpexpect.Expect, enabled bool) {
	r := e.PUT("/set").Expect().Status(http.StatusNoContent)

	r.Cookies().ContainsOnly("myname")
	c := r.Cookie("myname")
	c.Value().IsEqual("myvalue")
	c.Path().IsEqual("/")
	c.Expires().IsEqual(time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC))

	if enabled {
		e.GET("/get").Expect().Status(http.StatusOK).Text().IsEqual("myvalue")
	} else {
		e.GET("/get").Expect().Status(http.StatusBadRequest)
	}
}

func TestE2ECookie_LiveDisabled(t *testing.T) {
	handler := createCookieHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Jar: nil,
		},
	})

	testCookieHandler(e, false)
}

func TestE2ECookie_LiveEnabled(t *testing.T) {
	handler := createCookieHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Jar: httpexpect.NewCookieJar(),
		},
	})

	testCookieHandler(e, true)
}

func TestE2ECookie_BinderStandardDisabled(t *testing.T) {
	handler := createCookieHandler()

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       nil,
		},
	})

	testCookieHandler(e, false)
}

func TestE2ECookie_BinderStandardEnabled(t *testing.T) {
	handler := createCookieHandler()

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewCookieJar(),
		},
	})

	testCookieHandler(e, true)
}

func TestE2ECookie_BinderFastDisabled(t *testing.T) {
	handler := fasthttpadaptor.NewFastHTTPHandler(createCookieHandler())

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       nil,
		},
	})

	testCookieHandler(e, false)
}

func TestE2ECookie_BinderFastEnabled(t *testing.T) {
	handler := fasthttpadaptor.NewFastHTTPHandler(createCookieHandler())

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       httpexpect.NewCookieJar(),
		},
	})

	testCookieHandler(e, true)
}
