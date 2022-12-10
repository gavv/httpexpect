package httpexpect

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCookieFailed(t *testing.T) {
	check := func(value *Cookie, isNil bool) {
		value.chain.assertFailed(t)

		if isNil {
			assert.Nil(t, value.Raw())
		} else {
			assert.NotNil(t, value.Raw())
		}
		assert.NotNil(t, value.Name())
		assert.NotNil(t, value.Value())
		assert.NotNil(t, value.Domain())
		assert.NotNil(t, value.Path())
		assert.NotNil(t, value.Expires())
		assert.NotNil(t, value.MaxAge())

		value.HaveMaxAge()
		value.NotHaveMaxAge()
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(mockFailure())

		value := newCookie(chain, &http.Cookie{})

		check(value, false)
	})

	t.Run("nil_value", func(t *testing.T) {
		chain := newMockChain(t)

		value := newCookie(chain, nil)

		check(value, true)
	})

	t.Run("failed_chain_nil_value", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(mockFailure())

		value := newCookie(chain, nil)

		check(value, true)
	})
}

func TestCookieGetters(t *testing.T) {
	reporter := newMockReporter(t)

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
	assert.Equal(t, 123*time.Second, value.MaxAge().Raw())

	value.chain.assertOK(t)
}

func TestCookieMaxAge(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("unset", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: 0,
		})

		value.chain.assertOK(t)

		value.HaveMaxAge().chain.assertFailed(t)
		value.chain.reset()

		value.NotHaveMaxAge().chain.assertOK(t)
		value.chain.reset()

		require.Nil(t, value.MaxAge().value)

		value.MaxAge().NotSet().chain.assertOK(t)
		value.MaxAge().IsSet().chain.assertFailed(t)
	})

	t.Run("zero", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: -1,
		})

		value.chain.assertOK(t)

		value.HaveMaxAge().chain.assertOK(t)
		value.chain.reset()

		value.NotHaveMaxAge().chain.assertFailed(t)
		value.chain.reset()

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

		value.HaveMaxAge().chain.assertOK(t)
		value.chain.reset()

		value.NotHaveMaxAge().chain.assertFailed(t)
		value.chain.reset()

		require.NotNil(t, value.MaxAge().value)
		require.Equal(t, 3*time.Second, *value.MaxAge().value)

		value.MaxAge().IsSet().chain.assertOK(t)
		value.MaxAge().Equal(3 * time.Second).chain.assertOK(t)
	})
}
