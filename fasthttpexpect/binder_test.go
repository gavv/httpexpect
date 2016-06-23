package fasthttpexpect

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestBinder(t *testing.T) {
	binder := NewBinder(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "GET", string(ctx.Request.Header.Method()))
		assert.Equal(t, "http://example.com", string(ctx.Request.Header.RequestURI()))
		assert.Equal(t, "body", string(ctx.Request.Body()))

		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte(`{"hello":"world"}`))
	})

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
