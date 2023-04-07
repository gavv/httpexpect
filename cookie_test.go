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
		value.chain.assert(t, failure)

		if isNil {
			assert.Nil(t, value.Raw())
		} else {
			assert.NotNil(t, value.Raw())
		}

		value.Alias("foo")

		value.Name().chain.assert(t, failure)
		value.Value().chain.assert(t, failure)
		value.Domain().chain.assert(t, failure)
		value.Path().chain.assert(t, failure)
		value.Expires().chain.assert(t, failure)
		value.MaxAge().chain.assert(t, failure)

		value.HasMaxAge()
		value.NotHasMaxAge()
	}

	t.Run("failed chain", func(t *testing.T) {
		chain := newMockChain(t, flagFailed)
		value := newCookie(chain, &http.Cookie{})

		check(value, false)
	})

	t.Run("nil value", func(t *testing.T) {
		chain := newMockChain(t)
		value := newCookie(chain, nil)

		check(value, true)
	})

	t.Run("failed chain, nil value", func(t *testing.T) {
		chain := newMockChain(t, flagFailed)
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
		value.chain.assert(t, success)
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
		value.chain.assert(t, success)
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

	value.chain.assert(t, success)

	value.Name().chain.assert(t, success)
	value.Value().chain.assert(t, success)
	value.Domain().chain.assert(t, success)
	value.Path().chain.assert(t, success)
	value.Expires().chain.assert(t, success)
	value.MaxAge().chain.assert(t, success)

	assert.Equal(t, "name", value.Name().Raw())
	assert.Equal(t, "value", value.Value().Raw())
	assert.Equal(t, "example.com", value.Domain().Raw())
	assert.Equal(t, "/path", value.Path().Raw())
	assert.True(t, time.Unix(1234, 0).Equal(value.Expires().Raw()))
	assert.Equal(t, 123*time.Second, value.MaxAge().Raw())

	value.chain.assert(t, success)
}

func TestCookie_MaxAge(t *testing.T) {
	cases := []struct {
		name          string
		maxAge        int
		wantHasMaxAge chainResult
		wantDuration  time.Duration
	}{
		{
			name:          "unset",
			maxAge:        0,
			wantHasMaxAge: failure,
		},
		{
			name:          "zero",
			maxAge:        -1,
			wantHasMaxAge: success,
			wantDuration:  time.Duration(0),
		},
		{
			name:          "non-zero",
			maxAge:        3,
			wantHasMaxAge: success,
			wantDuration:  time.Duration(3) * time.Second,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			data := &http.Cookie{
				MaxAge: tc.maxAge,
			}

			NewCookie(reporter, data).HasMaxAge().
				chain.assert(t, tc.wantHasMaxAge)

			NewCookie(reporter, data).NotHasMaxAge().
				chain.assert(t, !tc.wantHasMaxAge)

			if tc.wantHasMaxAge {
				require.NotNil(t,
					NewCookie(reporter, data).MaxAge().value)

				require.Equal(t, tc.wantDuration,
					*NewCookie(reporter, data).MaxAge().value)
			} else {
				require.Nil(t,
					NewCookie(reporter, data).MaxAge().value)
			}
		})
	}
}
