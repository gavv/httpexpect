package httpexpect

import (
	"encoding/json"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

var (
	log = logging.MustGetLogger("httpexpect")
)

type Expect struct {
	testing *testing.T
	address string
}

func New(t *testing.T, address string) *Expect {
	return &Expect{
		testing: t,
		address: address,
	}
}

func (e *Expect) HEAD(url string) *Response {
	return e.Request("HEAD", url, nil)
}

func (e *Expect) GET(url string) *Response {
	return e.Request("GET", url, nil)
}

func (e *Expect) POST(url string, payload interface{}) *Response {
	return e.Request("POST", url, payload)
}

func (e *Expect) PUT(url string, payload interface{}) *Response {
	return e.Request("PUT", url, payload)
}

func (e *Expect) DELETE(url string) *Response {
	return e.Request("DELETE", url, nil)
}

func (e *Expect) Request(method, url string, payload interface{}) *Response {
	url = joinUrl(e.address, url)

	log.Debugf("%s %s", method, url)

	var str string
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			require.FailNow(e.testing, err.Error())
		}
		str = string(b)
	}

	req, err := http.NewRequest(method, url, strings.NewReader(str))
	if err != nil {
		require.FailNow(e.testing, err.Error())
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		require.FailNow(e.testing, err.Error())
	}

	return &Response{e.testing, resp}
}

func joinUrl(a, b string) string {
	if strings.HasSuffix(a, "/") {
		a = a[:len(a)-1]
	}
	if strings.HasPrefix(b, "/") {
		b = b[1:len(b)]
	}
	return a + "/" + b
}
