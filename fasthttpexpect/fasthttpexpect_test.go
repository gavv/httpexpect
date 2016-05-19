package fasthttpexpect

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"testing"
)

type testClient interface {
	Do(*http.Request) (*http.Response, error)
}

type mockBackend struct {
	t *testing.T
}

func (c mockBackend) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	assert.Equal(c.t, "GET", string(req.Header.Method()))
	assert.Equal(c.t, "http://example.com", string(req.Header.RequestURI()))
	assert.Equal(c.t, "body", string(req.Body()))

	resp.Header.Set("Content-Type", "application/json")
	resp.SetBody([]byte(`{"hello":"world"}`))

	return nil
}

func runTest(t *testing.T, client testClient) {
	req, err := http.NewRequest(
		"GET", "http://example.com", bytes.NewReader([]byte("body")))

	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
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

func TestClientAdapter(t *testing.T) {
	adapter := WithClient(mockBackend{t})

	runTest(t, adapter)
}

func TestBinder(t *testing.T) {
	backend := mockBackend{t}

	binder := NewBinder(func(ctx *fasthttp.RequestCtx) {
		backend.Do(&ctx.Request, &ctx.Response)
	})

	runTest(t, binder)
}
