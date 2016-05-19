package httpexpect

import (
	"bytes"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
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
