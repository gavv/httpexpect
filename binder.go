package httpexpect

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// Binder implements networkless Client attached directly to http.Handler.
//
// Binder emulates network communication by using given http.Handler directly.
// It passes httptest.ResponseRecorder as http.ResponseWriter to the handler,
// and then construct http.Response from recorded data.
type Binder struct {
	handler http.Handler
}

// NewBinder returns a new Binder given http.Handler.
func NewBinder(handler http.Handler) *Binder {
	return &Binder{handler}
}

// Do implements Client.Do.
func (binder *Binder) Do(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()

	binder.handler.ServeHTTP(recorder, req)

	resp := http.Response{
		Request:    req,
		StatusCode: recorder.Code,
		Status:     http.StatusText(recorder.Code),
		Header:     recorder.HeaderMap,
	}

	if recorder.Body != nil {
		resp.Body = ioutil.NopCloser(recorder.Body)
	}

	return &resp, nil
}
