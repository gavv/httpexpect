package httpexpect

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func testCookieHandler(e *Expect, enabled bool) {
	r := e.PUT("/set").Expect().Status(http.StatusNoContent)

	r.Cookies().ContainsOnly("myname")
	c := r.Cookie("myname")
	c.Value().Equal("myvalue")
	c.Path().Equal("/")
	c.Expires().Equal(time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC))

	if enabled {
		e.GET("/get").Expect().Status(http.StatusOK).Text().Equal("myvalue")
	} else {
		e.GET("/get").Expect().Status(http.StatusBadRequest)
	}
}

func TestE2ECookie_LiveDisabled(t *testing.T) {
	handler := createCookieHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
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

	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Jar: NewCookieJar(),
		},
	})

	testCookieHandler(e, true)
}

func TestE2ECookie_BinderStandardDisabled(t *testing.T) {
	handler := createCookieHandler()

	e := WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewBinder(handler),
			Jar:       nil,
		},
	})

	testCookieHandler(e, false)
}

func TestE2ECookie_BinderStandardEnabled(t *testing.T) {
	handler := createCookieHandler()

	e := WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewBinder(handler),
			Jar:       NewCookieJar(),
		},
	})

	testCookieHandler(e, true)
}

func TestE2ECookie_BinderFastDisabled(t *testing.T) {
	handler := fasthttpadaptor.NewFastHTTPHandler(createCookieHandler())

	e := WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewFastBinder(handler),
			Jar:       nil,
		},
	})

	testCookieHandler(e, false)
}

func TestE2ECookie_BinderFastEnabled(t *testing.T) {
	handler := fasthttpadaptor.NewFastHTTPHandler(createCookieHandler())

	e := WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewFastBinder(handler),
			Jar:       NewCookieJar(),
		},
	})

	testCookieHandler(e, true)
}
