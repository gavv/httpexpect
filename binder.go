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
	// Handler specifies the function to be invoked.
	Handler http.Handler

	// Jar specifies the cookie jar.
	// If Jar is nil, cookies are not sent in requests and ignored
	// in responses.
	Jar http.CookieJar
}

// NewBinder returns a new Binder given http.Handler.
// It uses DefaultJar() as cookie jar.
func NewBinder(handler http.Handler) *Binder {
	return &Binder{
		Handler: handler,
		Jar:     DefaultJar(),
	}
}

// Do implements Client.Do.
func (binder *Binder) Do(req *http.Request) (*http.Response, error) {
	if binder.Jar != nil {
		for _, cookie := range binder.Jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	recorder := httptest.NewRecorder()

	binder.Handler.ServeHTTP(recorder, req)

	resp := http.Response{
		Request:    req,
		StatusCode: recorder.Code,
		Status:     http.StatusText(recorder.Code),
		Header:     recorder.HeaderMap,
	}

	if recorder.Body != nil {
		resp.Body = ioutil.NopCloser(recorder.Body)
	}

	if binder.Jar != nil {
		if rc := resp.Cookies(); len(rc) > 0 {
			binder.Jar.SetCookies(req.URL, rc)
		}
	}

	return &resp, nil
}
