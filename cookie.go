package httpexpect

import (
	"errors"
	"net/http"
	"time"
)

// Cookie provides methods to inspect attached http.Cookie value.
type Cookie struct {
	chain *chain
	value *http.Cookie
}

// NewCookie returns a new Cookie object given a reporter used to report
// failures and cookie value to be inspected.
//
// reporter and value should not be nil.
//
// Example:
//
//	cookie := NewCookie(reporter, &http.Cookie{...})
//	cookie.Domain().Equal("example.com")
//	cookie.Path().Equal("/")
//	cookie.Expires().InRange(time.Now(), time.Now().Add(time.Hour * 24))
func NewCookie(reporter Reporter, value *http.Cookie) *Cookie {
	return newCookie(newDefaultChain("Cookie()", reporter), value)
}

func newCookie(parent *chain, val *http.Cookie) *Cookie {
	c := &Cookie{parent.clone(), nil}

	if val == nil {
		c.chain.fail(&AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{val},
			Errors: []error{
				errors.New("expected: non-nil cookie"),
			},
		})
	} else {
		c.value = val
	}

	return c
}

// Raw returns underlying http.Cookie value attached to Cookie.
// This is the value originally passed to NewCookie.
//
// Example:
//
//	cookie := NewCookie(t, c)
//	assert.Equal(t, c, cookie.Raw())
func (c *Cookie) Raw() *http.Cookie {
	return c.value
}

// Name returns a new String object that may be used to inspect
// cookie name.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.Name().Equal("session")
func (c *Cookie) Name() *String {
	c.chain.enter("Name()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newString(c.chain, "")
	}

	return newString(c.chain, c.value.Name)
}

// Value returns a new String object that may be used to inspect
// cookie value.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.Value().Equal("gH6z7Y")
func (c *Cookie) Value() *String {
	c.chain.enter("Value()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newString(c.chain, "")
	}

	return newString(c.chain, c.value.Value)
}

// Domain returns a new String object that may be used to inspect
// cookie domain.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.Domain().Equal("example.com")
func (c *Cookie) Domain() *String {
	c.chain.enter("Domain()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newString(c.chain, "")
	}

	return newString(c.chain, c.value.Domain)
}

// Path returns a new String object that may be used to inspect
// cookie path.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.Path().Equal("/foo")
func (c *Cookie) Path() *String {
	c.chain.enter("Path()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newString(c.chain, "")
	}

	return newString(c.chain, c.value.Path)
}

// Expires returns a new DateTime object that may be used to inspect
// cookie expiration date.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.Expires().InRange(time.Now(), time.Now().Add(time.Hour * 24))
func (c *Cookie) Expires() *DateTime {
	c.chain.enter("Expires()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newDateTime(c.chain, time.Unix(0, 0))
	}

	return newDateTime(c.chain, c.value.Expires)
}

// MaxAge returns a new Duration object that may be used to inspect
// cookie Max-age field.
//
// If MaxAge is not set, the returned Duration is unset. Whether a Duration
// is set or not can be checked using its IsSet and NotSet methods.
//
// If MaxAge is zero (which means delete cookie now), the returned Duration
// is set and equals to zero.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.MaxAge().IsSet()
//	cookie.MaxAge().InRange(time.Minute, time.Minute*10)
func (c *Cookie) MaxAge() *Duration {
	c.chain.enter("MaxAge()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newDuration(c.chain, nil)
	}

	if c.value.MaxAge == 0 {
		return newDuration(c.chain, nil)
	}
	if c.value.MaxAge < 0 {
		var zero time.Duration
		return newDuration(c.chain, &zero)
	}

	d := time.Duration(c.value.MaxAge) * time.Second
	return newDuration(c.chain, &d)
}
