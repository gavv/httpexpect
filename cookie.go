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

// NewCookie returns a new Cookie instance.
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
	return newCookie(newChainWithDefaults("Cookie()", reporter), value)
}

func newCookie(parent *chain, val *http.Cookie) *Cookie {
	c := &Cookie{parent.clone(), nil}

	if val == nil {
		c.chain.fail(AssertionFailure{
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

// Name returns a new String instance with cookie name.
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

// Value returns a new String instance with cookie value.
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

// Domain returns a new String instance with cookie domain.
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

// Path returns a new String instance with cookie path.
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

// Expires returns a new DateTime instance with cookie expiration date.
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

// HaveMaxAge succeeds if cookie has Max-Age field.
//
// In particular, if Max-Age is present and is zero (which means delete
// cookie now), method succeeds.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.HaveMaxAge()
func (c *Cookie) HaveMaxAge() *Cookie {
	c.chain.enter("HaveMaxAge()")
	defer c.chain.leave()

	if c.chain.failed() {
		return c
	}

	if c.value.MaxAge == 0 {
		c.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{c.value},
			Errors: []error{
				errors.New("expected: cookie has Max-Age field"),
			},
		})
	}

	return c
}

// NotHaveMaxAge succeeds if cookie does not have Max-Age field.
//
// In particular, if Max-Age is present and is zero (which means delete
// cookie now), method fails.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.NotHaveMaxAge()
func (c *Cookie) NotHaveMaxAge() *Cookie {
	c.chain.enter("NotHaveMaxAge()")
	defer c.chain.leave()

	if c.chain.failed() {
		return c
	}

	if c.value.MaxAge != 0 {
		c.chain.fail(AssertionFailure{
			Type:   AssertNotValid,
			Actual: &AssertionValue{c.value},
			Errors: []error{
				errors.New("expected: cookie does not have Max-Age field"),
			},
		})
	}

	return c
}

// MaxAge returns a new Duration instance with cookie Max-Age field.
//
// If Max-Age is not present, method fails.
//
// If Max-Age is present and is zero (which means delete cookie now),
// methods succeeds and the returned Duration is equal to zero.
//
// Example:
//
//	cookie := NewCookie(t, &http.Cookie{...})
//	cookie.HaveMaxAge()
//	cookie.MaxAge().InRange(time.Minute, time.Minute*10)
func (c *Cookie) MaxAge() *Duration {
	c.chain.enter("MaxAge()")
	defer c.chain.leave()

	if c.chain.failed() {
		return newDuration(c.chain, nil)
	}

	switch {
	case c.value.MaxAge == 0: // zero value means not present
		// TODO: after removing Duration.IsSet, add failure here (breaking change)
		_ = (*Duration).IsSet

		return newDuration(c.chain, nil)

	case c.value.MaxAge < 0: // negative value means present and zero
		age := time.Duration(0)
		return newDuration(c.chain, &age)

	default:
		age := time.Duration(c.value.MaxAge) * time.Second
		return newDuration(c.chain, &age)
	}
}
