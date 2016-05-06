package httpexpect

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Request struct {
	config  Config
	method  string
	url     string
	headers map[string]string
	body    io.Reader
}

func NewRequest(config Config, method, url string) *Request {
	return &Request{
		config: config,
		method: method,
		url:    url,
	}
}

func (r *Request) WithHeaders(headers map[string]string) *Request {
	r.headers = headers
	return r
}

func (r *Request) WithHeader(k, v string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[k] = v
	return r
}

func (r *Request) WithBody(reader io.Reader) *Request {
	r.body = reader
	return r
}

func (r *Request) WithBytes(b []byte) *Request {
	return r.WithBody(bytes.NewReader(b))
}

func (r *Request) WithJSON(obj interface{}) *Request {
	b, err := json.Marshal(obj)
	if err != nil {
		r.config.Checker.Fail(err.Error())
		return r
	}

	r.WithHeader("Content-Type", "application/json")
	r.WithBytes(b)

	return r
}

func (r *Request) Expect() *Response {
	resp := r.sendRequest()
	return NewResponse(r.config.Checker.Clone(), resp)
}

func (r *Request) sendRequest() *http.Response {
	if r.config.Checker.Failed() {
		return nil
	}

	if r.config.Logger != nil {
		r.config.Logger.LogRequest(r.method, r.url)
	}

	req, err := http.NewRequest(r.method, r.url, r.body)
	if err != nil {
		r.config.Checker.Fail(err.Error())
		return nil
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	resp, err := r.config.Client.Do(req)
	if err != nil {
		r.config.Checker.Fail(err.Error())
		return nil
	}

	return resp
}
