package httpexpect

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCookieFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	value := &Cookie{chain, nil}

	assert.True(t, value.Raw() == nil)
	assert.True(t, value.Name() != nil)
	assert.True(t, value.Value() != nil)
	assert.True(t, value.Domain() != nil)
	assert.True(t, value.Path() != nil)
	assert.True(t, value.Expires() != nil)
	assert.True(t, value.MaxAge() != nil)
}

func TestCookieGetters(t *testing.T) {
	reporter := newMockReporter(t)

	NewCookie(reporter, nil).chain.assertFailed(t)

	value := NewCookie(reporter, &http.Cookie{
		Name:    "name",
		Value:   "value",
		Domain:  "example.com",
		Path:    "/path",
		Expires: time.Unix(1234, 0),
		MaxAge:  123,
	})

	value.chain.assertOK(t)

	value.Name().chain.assertOK(t)
	value.Value().chain.assertOK(t)
	value.Domain().chain.assertOK(t)
	value.Path().chain.assertOK(t)
	value.Expires().chain.assertOK(t)
	value.MaxAge().chain.assertOK(t)

	assert.Equal(t, "name", value.Name().Raw())
	assert.Equal(t, "value", value.Value().Raw())
	assert.Equal(t, "example.com", value.Domain().Raw())
	assert.Equal(t, "/path", value.Path().Raw())
	assert.True(t, time.Unix(1234, 0).Equal(value.Expires().Raw()))
	assert.Equal(t, time.Duration(123*time.Second), value.MaxAge().Raw())

	value.chain.assertOK(t)
}

func TestCookieMaxAge(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("unset", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: 0,
		})

		value.chain.assertOK(t)

		require.Nil(t, value.MaxAge().value)

		value.MaxAge().NotSet().chain.assertOK(t)
		value.MaxAge().IsSet().chain.assertFailed(t)
	})

	t.Run("zero", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: -1,
		})

		value.chain.assertOK(t)

		require.NotNil(t, value.MaxAge().value)
		require.Equal(t, time.Duration(0), *value.MaxAge().value)

		value.MaxAge().IsSet().chain.assertOK(t)
		value.MaxAge().Equal(0).chain.assertOK(t)
	})

	t.Run("non-zero", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: 3,
		})

		value.chain.assertOK(t)

		require.NotNil(t, value.MaxAge().value)
		require.Equal(t, 3*time.Second, *value.MaxAge().value)

		value.MaxAge().IsSet().chain.assertOK(t)
		value.MaxAge().Equal(3 * time.Second).chain.assertOK(t)
	})
}
