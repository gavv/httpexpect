package httpexpect

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"testing"
)

type mockHandler struct {
	t *testing.T
}

func (c mockHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	assert.True(c.t, err == nil)

	assert.Equal(c.t, "GET", req.Method)
	assert.Equal(c.t, "http://example.com", req.URL.String())
	assert.Equal(c.t, "body", string(body))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(`{"hello":"world"}`))

	assert.True(c.t, err == nil)
}

func TestBinder(t *testing.T) {
	binder := NewBinder(mockHandler{t})

	req, err := http.NewRequest(
		"GET", "http://example.com", bytes.NewReader([]byte("body")))

	if err != nil {
		t.Fatal(err)
	}

	resp, err := binder.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	header := http.Header{
		"Content-Type": {"application/json"},
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, header, resp.Header)
	assert.Equal(t, `{"hello":"world"}`, string(b))
}

func TestFastBinder(t *testing.T) {
	binder := NewFastBinder(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "POST", string(ctx.Request.Header.Method()))
		assert.Equal(t, "http://example.com", string(ctx.Request.Header.RequestURI()))
		assert.Equal(t, "application/x-www-form-urlencoded",
			string(ctx.Request.Header.ContentType()))

		assert.Equal(t, "bar", string(ctx.FormValue("foo")))
		assert.Equal(t, "foo=bar", string(ctx.Request.Body()))

		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte(`{"hello":"world"}`))
	})

	req, err := http.NewRequest(
		"POST", "http://example.com", bytes.NewReader([]byte("foo=bar")))

	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := binder.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	header := http.Header{
		"Content-Type": {"application/json"},
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, header, resp.Header)
	assert.Equal(t, `{"hello":"world"}`, string(b))
}
