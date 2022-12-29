package httpexpect

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createTimeoutHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
	})

	mux.HandleFunc("/small", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`"`))
		_, _ = w.Write([]byte(randomString(10)))
		_, _ = w.Write([]byte(`"`))
	})

	mux.HandleFunc("/large", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`"`))
		_, _ = w.Write([]byte(randomString(1024 * 10)))
		_, _ = w.Write([]byte(`"`))
	})

	return mux
}

func TestE2ETimeout_DeadlineExpired(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	handler := createTimeoutHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	r := newMockReporter(t)

	e := WithConfig(Config{
		BaseURL:  server.URL,
		Reporter: r,
	})

	e.GET("/sleep").
		WithTimeout(10 * time.Millisecond).
		Expect()

	assert.True(t, r.reported)
}

func TestE2ETimeout_SmallBody(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	handler := createTimeoutHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := Default(t, server.URL)

	for i := 0; i < 100; i++ {
		e.GET("/small").
			WithTimeout(20 * time.Minute).
			Expect().
			Status(http.StatusOK).
			JSON().
			String()
	}
}

func TestE2ETimeout_LargeBody(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	handler := createTimeoutHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := Default(t, server.URL)

	for i := 0; i < 100; i++ {
		e.GET("/large").
			WithTimeout(20 * time.Minute).
			Expect().
			Status(http.StatusOK).
			JSON().
			String()
	}
}
