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

func TestCookieConstructors(t *testing.T) {
	cookie := &http.Cookie{
		Name:    "Test",
		Value:   "Test_val",
		Domain:  "example.com",
		Path:    "/path",
		Expires: time.Unix(1234, 0),
		MaxAge:  123,
	}

	t.Run("Constructor without config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewCookie(reporter, cookie)
		value.Name().Equal("Test")
		value.Value().Equal("Test_val")
		value.Domain().Equal("example.com")
		value.Expires().Equal(time.Unix(1234, 0))
		value.MaxAge().Equal(123 * time.Second)
		value.chain.assertNotFailed(t)
	})

	t.Run("Constructor with config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewCookieC(Config{
			Reporter: reporter,
		}, cookie)
		value.Name().Equal("Test")
		value.Value().Equal("Test_val")
		value.Domain().Equal("example.com")
		value.Expires().Equal(time.Unix(1234, 0))
		value.MaxAge().Equal(123 * time.Second)
		value.chain.assertNotFailed(t)
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

	value.chain.assertNotFailed(t)

	value.Name().chain.assertNotFailed(t)
	value.Value().chain.assertNotFailed(t)
	value.Domain().chain.assertNotFailed(t)
	value.Path().chain.assertNotFailed(t)
	value.Expires().chain.assertNotFailed(t)
	value.MaxAge().chain.assertNotFailed(t)

	assert.Equal(t, "name", value.Name().Raw())
	assert.Equal(t, "value", value.Value().Raw())
	assert.Equal(t, "example.com", value.Domain().Raw())
	assert.Equal(t, "/path", value.Path().Raw())
	assert.True(t, time.Unix(1234, 0).Equal(value.Expires().Raw()))
	assert.Equal(t, 123*time.Second, value.MaxAge().Raw())

	value.chain.assertNotFailed(t)
}

func TestCookieMaxAge(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("unset", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: 0,
		})

		value.chain.assertNotFailed(t)

		value.HaveMaxAge().chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotHaveMaxAge().chain.assertNotFailed(t)
		value.chain.clearFailed()

		require.Nil(t, value.MaxAge().value)

		value.MaxAge().NotSet().chain.assertNotFailed(t)
		value.MaxAge().IsSet().chain.assertFailed(t)
	})

	t.Run("zero", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: -1,
		})

		value.chain.assertNotFailed(t)

		value.HaveMaxAge().chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotHaveMaxAge().chain.assertFailed(t)
		value.chain.clearFailed()

		require.NotNil(t, value.MaxAge().value)
		require.Equal(t, time.Duration(0), *value.MaxAge().value)

		value.MaxAge().IsSet().chain.assertNotFailed(t)
		value.MaxAge().Equal(0).chain.assertNotFailed(t)
	})

	t.Run("non-zero", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: 3,
		})

		value.chain.assertNotFailed(t)

		value.HaveMaxAge().chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotHaveMaxAge().chain.assertFailed(t)
		value.chain.clearFailed()

		require.NotNil(t, value.MaxAge().value)
		require.Equal(t, 3*time.Second, *value.MaxAge().value)

		value.MaxAge().IsSet().chain.assertNotFailed(t)
		value.MaxAge().Equal(3 * time.Second).chain.assertNotFailed(t)
	})
}
