package httpexpect

import (
	"net/http/cookiejar"
	"net/http"

	"golang.org/x/net/publicsuffix"
)

// NewJar returns a new http.CookieJar.
//
// Returned jar is implemented in net/http/cookiejar. PublicSuffixList is
// implemented in golang.org/x/net/publicsuffix.
//
// Note that this jar ignores cookies when request url is empty.
func NewJar() http.CookieJar {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		panic(err)
	}
	return jar
}
