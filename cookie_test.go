package httpexpect

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCookie_FailedChain(t *testing.T) {
	check := func(value *Cookie, isNil bool) {
		value.chain.assertFailed(t)

		if isNil {
			assert.Nil(t, value.Raw())
		} else {
			assert.NotNil(t, value.Raw())
		}

		value.Alias("foo")

		value.Name().chain.assertFailed(t)
		value.Value().chain.assertFailed(t)
		value.Domain().chain.assertFailed(t)
		value.Path().chain.assertFailed(t)
		value.Expires().chain.assertFailed(t)
		value.MaxAge().chain.assertFailed(t)

		value.HasMaxAge()
		value.NotHasMaxAge()
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.setFailed()

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
		chain.setFailed()

		value := newCookie(chain, nil)

		check(value, true)
	})
}

func TestCookie_Constructors(t *testing.T) {
	cookie := &http.Cookie{
		Name:    "Test",
		Value:   "Test_val",
		Domain:  "example.com",
		Path:    "/path",
		Expires: time.Unix(1234, 0),
		MaxAge:  123,
	}

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewCookie(reporter, cookie)
		value.Name().IsEqual("Test")
		value.Value().IsEqual("Test_val")
		value.Domain().IsEqual("example.com")
		value.Expires().IsEqual(time.Unix(1234, 0))
		value.MaxAge().IsEqual(123 * time.Second)
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewCookieC(Config{
			Reporter: reporter,
		}, cookie)
		value.Name().IsEqual("Test")
		value.Value().IsEqual("Test_val")
		value.Domain().IsEqual("example.com")
		value.Expires().IsEqual(time.Unix(1234, 0))
		value.MaxAge().IsEqual(123 * time.Second)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newCookie(chain, cookie)
		assert.NotSame(t, value.chain, &chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestCookie_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewCookie(reporter, &http.Cookie{
		MaxAge: 0,
	})
	assert.Equal(t, []string{"Cookie()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Cookie()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Cookie()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.Domain()
	assert.Equal(t, []string{"Cookie()", "Domain()"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Domain()"}, childValue.chain.context.AliasedPath)
}

func TestCookie_Getters(t *testing.T) {
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

func TestCookie_MaxAge(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("unset", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: 0,
		})

		value.chain.assertNotFailed(t)

		value.HasMaxAge().chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotHasMaxAge().chain.assertNotFailed(t)
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

		value.HasMaxAge().chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotHasMaxAge().chain.assertFailed(t)
		value.chain.clearFailed()

		require.NotNil(t, value.MaxAge().value)
		require.Equal(t, time.Duration(0), *value.MaxAge().value)

		value.MaxAge().IsSet().chain.assertNotFailed(t)
		value.MaxAge().IsEqual(0).chain.assertNotFailed(t)
	})

	t.Run("non-zero", func(t *testing.T) {
		value := NewCookie(reporter, &http.Cookie{
			MaxAge: 3,
		})

		value.chain.assertNotFailed(t)

		value.HasMaxAge().chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotHasMaxAge().chain.assertFailed(t)
		value.chain.clearFailed()

		require.NotNil(t, value.MaxAge().value)
		require.Equal(t, 3*time.Second, *value.MaxAge().value)

		value.MaxAge().IsSet().chain.assertNotFailed(t)
		value.MaxAge().IsEqual(3 * time.Second).chain.assertNotFailed(t)
	})
}
