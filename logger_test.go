package httpexpect

import (
	"bytes"
	"net/http"
	"testing"
)

func TestCompactLogger(t *testing.T) {
	logger := NewCompactLogger(t)

	body1 := bytes.NewBufferString("body1")
	body2 := bytes.NewBufferString("body2")

	req1, _ := http.NewRequest("GET", "http://example.com", body1)
	req2, _ := http.NewRequest("GET", "http://example.com", nil)

	logger.Request(req1)
	logger.Request(req2)
	logger.Request(nil)

	logger.Response(&http.Response{Body: closingBuffer{body2}})
	logger.Response(&http.Response{})
	logger.Response(nil)
}

func TestDebugLogger(t *testing.T) {
	logger := NewDebugLogger(t, true)

	body1 := bytes.NewBufferString("body1")
	body2 := bytes.NewBufferString("body2")

	req1, _ := http.NewRequest("GET", "http://example.com", body1)
	req2, _ := http.NewRequest("GET", "http://example.com", nil)

	logger.Request(req1)
	logger.Request(req2)
	logger.Request(nil)

	logger.Response(&http.Response{Body: closingBuffer{body2}})
	logger.Response(&http.Response{})
	logger.Response(nil)
}
